#!/usr/bin/env python

import socket
import io
import os
import random
import string
import datetime
import subprocess

import MySQLdb.cursors
import flask
import bcrypt
import pathlib
import requests

base_path = pathlib.Path(__file__).resolve().parent.parent
static_folder = base_path / 'public'

app = flask.Flask(__name__, static_folder=str(static_folder), static_url_path='')
app.config['SECRET_KEY'] = 'isucari'


class Constants(object):
    ITEM_STATUS_ON_SALE = "on_sale"
    ITEM_STATUS_TRADING = 'trading'
    ITEM_STATUS_SOLD_OUT = 'sold_out'
    ITEM_STATUS_STOP = 'stop'
    ITEM_STATUS_CANCEL = 'cancel'
    TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING = 'wait_shipping'
    TRANSACTION_EVIDENCE_STATUS_WAIT_DONE = 'wait_done'
    TRANSACTION_EVIDENCE_STATUS_DONE = 'done'

    SHIPPING_STATUS_INITIAL = 'initial'
    SHIPPING_STATUS_WAIT_PICKUP = 'wait_pickup'
    SHIPPING_STATUS_SHIPPING = 'shipping'
    SHIPPING_STATUS_DONE = 'done'

    ISUCARI_API_TOKEN = 'Bearer 75ugk2m37a750fwir5xr-22l6h4wmue1bwrubzwd0'

    PAYMENT_SERVICE_ISUCARI_API_KEY = 'a15400e46c83635eb181-946abb51ff26a868317c'
    PAYMENT_SERVICE_ISUCARI_SHOP_ID = '11'

    ITEMS_PER_PAGE = 48
    TRANSACTIONS_PER_PAGE = 10

def dbh():
    if hasattr(flask.g, 'db'):
        return flask.g.db

    flask.g.db = MySQLdb.connect(
        host=os.getenv('MYSQL_HOST', '127.0.0.1'),
        port=os.getenv('MYSQL_PORT', 3306),
        user=os.getenv('MYSQL_USER', 'isucari'),
        password=os.getenv('MYSQL_PASS', 'isucari'),
        db=os.getenv('MYSQL_DBNAME', 'isucari'),
        charset='utf8mb4',
        cursorclass=MySQLdb.cursors.DictCursor,
        autocommit=True,
    )
    cur = flask.g.db.cursor()
    cur.execute(
        "SET SESSION sql_mode='STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION'")
    return flask.g.db


def http_json_error(code, msg):
    flask.abort(flask.jsonify(code, {'error': msg}))


def random_string(length):
    letters = string.ascii_lowercase + string.digits
    return ''.join(random.choice(letters) for _ in range(length))


def get_user():
    user_id = flask.session.get("user_id")
    if user_id is None:
        http_json_error(requests.codes['not_found'], "no session")
    try:
        conn = dbh()
        with conn.cursor() as c:
            sql = "SELECT * FROM `users` WHERE `id` = %s"
            c.execute(sql, [user_id])
            user = c.fetchone()
            if user is None:
                http_json_error(requests.codes['not_found'], "user not found")
    except MySQLdb.Error as err:
        app.logger.exception(err)
        http_json_error(requests.codes['internal_server_error'], "db error")
    return user

def get_user_simple_by_id(user_id):
    try:
        conn = dbh()
        with conn.cursor() as c:
            sql = "SELECT * FROM `users` WHERE `id` = %s"
            c.execute(sql, [user_id])
            user = c.fetchone()
            if user is None:
                http_json_error(requests.codes['not_found'], "user not found")
    except MySQLdb.Error as err:
        app.logger.exception(err)
        http_json_error(requests.codes['internal_server_error'], "db error")
    return user

def get_category_by_id(category_id):
    conn = dbh()
    sql = "SELECT * FROM `categories` WHERE `id` = %s"
    with conn.cursor() as c:
        c.execute(sql, (category_id,))
        category = c.fetchone()
        # TODO: check err
    if category['parent_id'] != 0:
        parent = get_category_by_id(category['parent_id'])
        print(parent)
        if parent is not None:
            category['parent_category_name'] = parent['category_name']
    return category


def to_user_json(user):
    del (user['hashed_password'], user['last_bump'], user['created_at'])
    return user

def to_item_json(item):
    item["created_at"] = int(item["created_at"].timestamp())
    item["updated_at"] = int(item["updated_at"].timestamp())
    return item

def ensure_required_payload(keys=None):
    if keys is None:
        keys = []
    for k in keys:
        if k not in flask.request.json or len(flask.request.json[k]) == 0:
            http_json_error(requests.codes['bad_request'], 'all parameters are required')


def ensure_valid_csrf_token():
    if flask.request.json['csrf_token'] != flask.session['csrf_token']:
        http_json_error(requests.codes['unprocessable_entity'], "csrf token error")


def get_config(name):
    conn = dbh()
    sql = "SELECT * FROM `configs` WHERE `name` = ?"
    with conn.cursor() as c:
        c.execute(sql, name)
        config = c.fetchone()
    return config


def get_shipment_service_url():
    config = get_config("shipment_service_url")
    if config is None:
        return "http://localhost:7000"
    return config['val']


def get_payment_service_url():
    config = get_config("payment_service_url")
    if config is None:
        return "http://localhost:5000"
    return config['val']


# API
@app.route("/initialize", methods=["POST"])
def post_initialize():
    subprocess.call(["../sql/init.sh"])
    return ('', 204)


@app.route("/new_items.json", methods=["GET"])
def get_new_items():
    # TODO: check err
    if request.args.get('item_id') is not None:
        item_id = int(request.args.get('item_id'))

        if item_id <= 0:
            http_json_error(requests.codes['bad_request'], 'item_id param error')

    if request.args.get('created_at') is not None:
        created_at = int(request.args.get('created_at'))

        if created_at <= 0:
            http_json_error(requests.codes['bad_request'], 'created_at param error')

    items = []

    try:
        conn = dbh()
        with conn.cursor() as c:
            if item_id > 0 and created_at > 0:
                # paging
                sql = "SELECT * FROM `items` WHERE `status` IN (?,?) AND `created_at` <= ? AND `id` < ? ORDER BY `created_at` DESC, `id` DESC LIMIT ?"
                c.execute(sql, (
                    Constants.ITEM_STATUS_ON_SALE,
                    Constants.ITEM_STATUS_SOLD_OUT,
                    created_at.timestamp(),
                    item_id,
                    int(Constants.ITEMS_PER_PAGE) + 1,
                ))
            else:
                # 1st page
                sql = "SELECT * FROM `items` WHERE `status` IN (?,?) ORDER BY `created_at` DESC, `id` DESC LIMIT ?"
                c.execute(sql, (
                    Constants.ITEM_STATUS_ON_SALE,
                    Constants.ITEM_STATUS_SOLD_OUT,
                    int(Constants.ITEMS_PER_PAGE) + 1
                ))

        item_simples = []

    except MySQL.db.Error as err:
        pass

    return


@app.route("/new_items/<root_category_id>.json", methods=["GET"])
def get_new_category_items(root_category_id=None):
    return


@app.route("/users/transactions.json", methods=["GET"])
def get_transactions():
    user = get_user()
    conn = dbh()

    item_id = 0 # FIXME:
    created_at = 0 # FIXME:

    with conn.cursor() as c:

        try:

            if item_id > 0 and created_at > 0:
                sql = "SELECT * FROM `items` WHERE (`seller_id` = %s OR `buyer_id` = %s) AND `status` IN (%s,%s,%s,%s,%s) AND `created_at` <= %s AND `id` < %s ORDER BY `created_at` DESC, `id` DESC LIMIT %s"
                c.execute(sql, (
                    user['id'],
                    user['id'],
        			Constants.ITEM_STATUS_ON_SALE,
        			Constants.ITEM_STATUS_TRADING,
        			Constants.ITEM_STATUS_SOLD_OUT,
        			Constants.ITEM_STATUS_CANCEL,
        			Constants.ITEM_STATUS_STOP,
        			datetime.datetime.fromtimestamp(created_at),
        			item_id,
                    Constants.TRANSACTIONS_PER_PAGE+1,
                ))

            else:
                sql = "SELECT * FROM `items` WHERE (`seller_id` = %s OR `buyer_id` = %s ) AND `status` IN (%s,%s,%s,%s,%s) ORDER BY `created_at` DESC, `id` DESC LIMIT %s"
                c.execute(sql, [
                    user['id'],
                    user['id'],
        			Constants.ITEM_STATUS_ON_SALE,
        			Constants.ITEM_STATUS_TRADING,
        			Constants.ITEM_STATUS_SOLD_OUT,
        			Constants.ITEM_STATUS_CANCEL,
        			Constants.ITEM_STATUS_STOP,
                    Constants.TRANSACTIONS_PER_PAGE+1,
                ])

            item_details = []
            while True:
                item = c.fetchone()

                if item is None:
                    break

                seller = get_user_simple_by_id(item["seller_id"])
                category = get_category_by_id(item["category_id"])

                item = to_item_json(item)
                item["category"] = category
                item["seller"] = to_user_json(seller)

                print(item)
                item_details.append(item)

        except MySQLdb.Error as err:
            app.logger.exception(err)
            http_json_error(requests.codes['internal_server_error'], "db error")

    has_next = False
    if len(item_details) > Constants.TRANSACTIONS_PER_PAGE:
        has_next = True
        item_details = item_details[:Constants.TRANSACTIONS_PER_PAGE]

    return flask.jsonify(dict(
        items=item_details,
        has_next=has_next,
    ))

@app.route("/users/<user_id>.json", methods=["GET"])
def get_user_items(user_id=None):
    user = get_user_simple_by_id(user_id)
    conn = dbh()

    item_id = 0 # FIXME:
    created_at = 0 # FIXME:


    with conn.cursor() as c:

        try:

            if item_id > 0 and created_at > 0:
                sql = "SELECT * FROM `items` WHERE `seller_id` = %s AND `status` IN (%s,%s,%s) AND `created_at` <= %s AND `id` < %s ORDER BY `created_at` DESC, `id` DESC LIMIT %s"
                c.execute(sql, (
                    user['id'],
        			Constants.ITEM_STATUS_ON_SALE,
        			Constants.ITEM_STATUS_TRADING,
        			Constants.ITEM_STATUS_SOLD_OUT,
        			datetime.datetime.fromtimestamp(created_at),
        			item_id,
                    Constants.ITEMS_PER_PAGE+1,
                ))

            else:
                sql = "SELECT * FROM `items` WHERE `seller_id` = %s AND `status` IN (%s,%s,%s) ORDER BY `created_at` DESC, `id` DESC LIMIT %s"
                c.execute(sql, (
                    user['id'],
        			Constants.ITEM_STATUS_ON_SALE,
        			Constants.ITEM_STATUS_TRADING,
        			Constants.ITEM_STATUS_SOLD_OUT,
                    Constants.TRANSACTIONS_PER_PAGE+1,
                ))

            item_simples = []
            while True:
                item = c.fetchone()

                if item is None:
                    break

                seller = get_user_simple_by_id(item["seller_id"])
                category = get_category_by_id(item["category_id"])

                item = to_item_json(item)
                item["category"] = category
                item["seller"] = to_user_json(seller)

                print(item)
                item_simples.append(item)

        except MySQLdb.Error as err:
            app.logger.exception(err)
            http_json_error(requests.codes['internal_server_error'], "db error")

    has_next = False
    if len(item_simples) > Constants.ITEMS_PER_PAGE:
        has_next = True
        item_simples = item_simples[:Constants.ITEMS_PER_PAGE]

    return flask.jsonify(dict(
        user=to_user_json(user),
        items=item_simples,
        has_next=has_next,
    ))


@app.route("/items/<item_id>.json", methods=["GET"])
def get_item(item_id=None):
    return


@app.route("/itemds/edit", methods=["POST"])
def post_item_edit():
    ensure_valid_csrf_token()
    ensure_required_payload(['price'])

    price = int(flask.request.json['price'])
    if not 100 <= price <= 100000:
        http_json_error(requests.codes['bad_request'], "商品価格は100ｲｽｺｲﾝ以上、1,000,000ｲｽｺｲﾝ以下にしてください")
    user = get_user()
    conn = dbh()
    with conn.cursor() as c:
        try:
            sql = "SELECT * FROM `items` WHERE `id` = ?"
            c.execute(sql, user['id'])
            item = c.fetchone()
            if item is None:
                http_json_error(requests.codes['not_found'], "item not found")
            if item["seller_id"] != user["id"]:
                http_json_error(requests.codes['forbidden'], "自分の商品以外は編集できません")
        except MySQLdb.Error as err:
            app.logger.exception(err)
            http_json_error(requests.codes['internal_server_error'], "db error")

    conn.begin()
    with conn.cursor() as c:
        try:
            sql = "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE"
            c.execute(sql, flask.request.json["item_id"])
            item = c.fetchone()
            if item["status"] != Constants.ITEM_STATUS_ON_SALE:
                conn.rollback()
                http_json_error(requests.codes['forbidden'], "販売中の商品以外編集できません")
            sql = "UPDATE `items` SET `price` = ?, `updated_at` = ? WHERE `id` = ?"
            c.execute(sql, (
                flask.request.json["price"],
                datetime.datetime.now(),
                flask.request.json["item_id"]
            ))

            sql = "SELECT * FROM `items` WHERE `id` = ?"
            c.execute(sql, flask.request.json["item_id"])
            item = c.fetchone()
            conn.commit()
        except MySQLdb.Error as err:
            conn.rollback()
            app.logger.exception(err)
            http_json_error(requests.codes['internal_server_error'], "db error")
    return flask.jsonify(dict(
        item_id=item["id"],
        item_price=item["price"],
        item_created_at=int(item["created_at"].timestamp()),
        item_updated_at=int(item["updated_at"].timestamp()),
    ))


@app.route("/buy", methods=["POST"])
def post_buy():
    ensure_valid_csrf_token()
    buyer = get_user()

    conn = dbh()
    try:
        conn.begin()
        with conn.cursor() as c:
            sql = "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE"
            c.execute(sql, (flask.request.json['item_id']))
            item = c.fetchone()
            if item is None:
                conn.rollback()
                http_json_error(requests.codes['not_found'], "item not found")
            if item['status'] is not Constants.ITEM_STATUS_ON_SALE:
                conn.rollback()
                http_json_error(requests.codes['forbidden'], "item is not for sale")
            if item['seller_id'] == buyer['id']:
                conn.rollback()
                http_json_error(requests.codes['forbidden'], "自分の商品は買えません")
            sql = "SELECT * FROM `users` WHERE `id` = ? FOR UPDATE"
            c.execute(sql, (item['seller_id']))
            seller = c.fetchone()
            if seller is None:
                conn.rollback()
                http_json_error(requests.codes['not_found'], "seller not found")
            category = get_category_by_id(item['category_id'])
            # TODO: check category error
            sql = "INSERT INTO `transaction_evidences` (`seller_id`, `buyer_id`, `status`, `item_id`, `item_name`, " \
                  "`item_price`, `item_description`, `item_category_id`, `item_root_category_id`) " \
                  "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
            c.execute(sql, (
                item['seller_id'],
                buyer['id'],
                Constants.TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING,
                item['id'],
                item['name'],
                item['price'],
                item['description'],
                category['id'],
                category['parent_id'],
            ))
            transaction_evidence_id = c.lastrowid()
            sql = "UPDATE `items` SET `buyer_id` = ?, `status` = ?, `updated_at` = ? WHERE `id` = ?"
            c.execute(sql, (
                buyer['id'],
                Constants.ITEM_STATUS_TRADING,
                datetime.datetime.now(),
                item['id'],
            ))

            host = get_shipment_service_url()
            try:
                res = requests.post(host + "/create",
                                    headers=dict(Authorization=Constants.ISUCARI_API_TOKEN),
                                    json=dict(
                                        to_address=buyer['address'],
                                        to_name=buyer['account_name'],
                                        from_address=seller['address'],
                                        from_name=seller['name'],
                                    ))
                res.raise_for_status()
            except (socket.gaierror, requests.HTTPError) as err:
                conn.rollback()
                app.logger.exception(err)
                http_json_error(requests.codes['internal_server_error'])

            shipping_res = res.json()

            host = get_payment_service_url()
            try:
                res = requests.post(host + "/token",
                                    json=dict(
                                        shop_id=Constants.PAYMENT_SERVICE_ISUCARI_SHOP_ID,
                                        api_key=Constants.PAYMENT_SERVICE_ISUCARI_API_KEY,
                                        token=flask.request.json['token'],
                                        price=item['price'],
                                    ))
                res.raise_for_status()
            except (socket.gaierror, requests.HTTPError) as err:
                conn.rollback()
                app.logger.exception(err)
                http_json_error(requests.codes['internal_server_error'])

            payment_res = res.json()
            if payment_res['status'] == "invalid":
                conn.rollback()
                http_json_error(requests.codes["bad_request"], "カード情報に誤りがあります")
            if payment_res['status'] == "fail":
                conn.rollback()
                http_json_error(requests.codes["bad_request"], "カードの残高が足りません")
            if payment_res['status'] != "ok":
                conn.rollback()
                http_json_error(requests.codes["bad_request"], "想定外のエラー")

            sql = "INSERT INTO `shippings` (`transaction_evidence_id`, `status`, `item_name`, `item_id`, " \
                  "`reserve_id`, `reserve_time`, `to_address`, `to_name`, `from_address`, `from_name`, `img_binary`) " \
                  "VALUES (?,?,?,?,?,?,?,?,?,?,?) "
            c.execute(sql, (
                transaction_evidence_id,
                Constants.SHIPPING_STATUS_INITIAL,
                item["name"],
                item["id"],
                shipping_res["reserve_id"],
                shipping_res["reserve_time"],
                buyer["address"],
                buyer["account_name"],
                seller["address"],
                seller["account_name"],
                ""
            ))
        conn.commit()
    except MySQLdb.Error as err:
        app.logger.exception(err)
        http_json_error(requests.codes['internal_server_error'], "db error")
    return flask.jsonify(dict(transaction_evidence_id=transaction_evidence_id))


@app.route("/sell", methods=["POST"])
def post_sell():
    if flask.request.form['csrf_token'] != flask.session['csrf_token']:
        http_json_error(requests.codes['unprocessable_entity'], "csrf token error")
    for k in ["name", "description", "price", "category_id"]:
        if k not in flask.request.form or len(flask.request.form[k]) == 0:
            http_json_error(requests.codes['bad_request'], 'all parameters are required')

    price = int(flask.request.form['price'])
    if not 100 <= price <= 100000:
        http_json_error(requests.codes['bad_request'], "商品価格は100ｲｽｺｲﾝ以上、1,000,000ｲｽｺｲﾝ以下にしてください")

    category = get_category_by_id(flask.request.form['category_id'])
    if category['parent_category_id'] == 0:
        http_json_error(requests.codes['bad_request'], 'Incorrect category ID')
    user = get_user()
    if flask.request.files['image'] not in flask.request.files:
        http_json_error(requests.codes['internal_server_error'], 'image error')

    file = flask.request.files['image']
    ext = os.path.splitext(file.filename)[1]
    if ext not in ('.jpg', 'jpeg', '.png', 'gif'):
        http_json_error(requests.codes['bad_request'], 'unsupported image format error error')
    if ext == ".jpeg":
        ext = ".jpg"
    imagename = "{0}{1}".format(random_string(32), ext)
    file.save(os.path.join(app.config['UPLOAD_FOLDER'], imagename))

    try:
        conn = dbh()
        conn.begin()
        sql = "SELECT * FROM `users` WHERE `id` = ? FOR UPDATE"
        with conn.cursor() as c:
            c.execute(sql, (user['id']))
            seller = c.fetchone()
            if seller is None:
                conn.rollback()
                http_json_error(requests['not_found'], 'user not found')
            sql = """INSERT INTO `items`
            (`seller_id`, `status`, `name`, `price`, `description`, `image_name`, `category_id`)
             VALUES (?, ?, ?, ?, ?, ?, ?)"""
            c.execute(sql, (
                seller['id'],
                Constants.ITEM_STATUS_ON_SALE,
                flask.request.form['name'],
                flask.request.form['price'],
                flask.request.form['description'],
                imagename,
                flask.request.form['category_id'],
            ))
            item_id = c.lastrowid
            sql = "UPDATE `users` SET `num_sell_items`=?, `last_bump`=? WHERE `id`=?"
            c.execute(sql, (seller['num_sell_items'] + 1, datetime.datetime.now()))
            conn.commit()
    except MySQLdb.Error as err:
        app.logger.exception(err)
        http_json_error(requests.codes['internal_server_error'], "db error")

    return flask.jsonify({
        'id': item_id,
    })


@app.route("/ship", methods=["POST"])
def post_ship():
    ensure_valid_csrf_token()
    user = get_user()
    conn = dbh()
    with conn.cursor() as c:
        try:
            sql = "SELECT * FROM `transaction_evidences` WHERE `item_id` = ?"
            c.execute(sql, flask.request.json["item_id"])
            transaction_evidence = c.fetchone()
            if transaction_evidence is None:
                http_json_error(requests.codes["not_found"], "transaction_evidences not found")
        except MySQLdb.Error as err:
            app.logger.exception(err)
            http_json_error(requests.codes['internal_server_error'], "db error")
    if transaction_evidence["seller_id"] != user["id"]:
        http_json_error(requests.codes['forbidden'], "権限がありません")

    try:
        conn.begin()
        with conn.cursor() as c:
            sql = "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE"
            c.execute(sql, flask.request.json["item_id"])
            item = c.fetchone()
            if item is None:
                conn.rollback()
                http_json_error(requests.codes["not_found"], "item not found")
            if item["status"] != Constants.ITEM_STATUS_TRADING:
                conn.rollback()
                http_json_error(requests.codes["forbidden"], "商品が取引中ではありません")

            sql = "SELECT * FROM `transaction_evidences` WHERE `id` = ? FOR UPDATE"
            c.execute(sql, transaction_evidence["id"])
            transaction_evidence = c.fetchone()
            if transaction_evidence is None:
                conn.rollback()
                http_json_error(requests.codes["not_found"], "transaction_evidences not found")
            if transaction_evidence["status"] != Constants.TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING:
                conn.rollback()
                http_json_error(requests.codes['forbidden'], "準備ができていません")

            sql = "SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE"
            c.execute(sql, transaction_evidence["id"])
            shipping = c.fetchone()
            if shipping is None:
                conn.rollback()
                http_json_error(requests.codes["not_found"], "shipping not found")

            try:
                host = get_shipment_service_url()
                res = requests.post(host + "/request",
                                    header=dict(Authorization=Constants.ISUCARI_API_TOKEN),
                                    json=dict(reserve_id=shipping["reserve_id"]))
                res.raise_for_status()
            except (socket.gaierror, requests.HTTPError) as err:
                conn.rollback()
                app.logger.exception(err)
                http_json_error(requests.codes["internal_server_error"], "failed to request to shipment service")

            sql = "UPDATE `shippings` SET `status` = ?, `img_binary` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?"
            c.execute(sql, (
                Constants.SHIPPING_STATUS_WAIT_PICKUP,
                io.BytesIO(res.content),
                datetime.datetime.now(),
                transaction_evidence["id"],
            ))
        conn.commit()
    except MySQLdb.Error as err:
        app.logger.exception(err)
        http_json_error(requests.codes['internal_server_error'], "db error")
    return flask.jsonify(dict(path="/transactions/%d.png".format(transaction_evidence["id"])))


@app.route("/ship_done", methods=["POST"])
def post_ship_done():
    ensure_valid_csrf_token()
    user = get_user()
    conn = dbh()
    with conn.cursor() as c:
        try:
            sql = "SELECT * FROM `transaction_evidences` WHERE `item_id` = ?"
            c.execute(sql, flask.request.json["item_id"])
            transaction_evidence = c.fetchone()
            if transaction_evidence is None:
                http_json_error(requests.codes["not_found"], "transaction_evidences not found")
        except MySQLdb.Error as err:
            app.logger.exception(err)
            http_json_error(requests.codes['internal_server_error'], "db error")
    if transaction_evidence["seller_id"] != user["id"]:
        http_json_error(requests.codes['forbidden'], "権限がありません")

    try:
        conn.begin()
        with conn.cursor() as c:
            sql = "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE"
            c.execute(sql, flask.request.json["item_id"])
            item = c.fetchone()
            if item is None:
                conn.rollback()
                http_json_error(requests.codes["not_found"], "item not found")
            if item["status"] != Constants.ITEM_STATUS_TRADING:
                conn.rollback()
                http_json_error(requests.codes["forbidden"], "商品が取引中ではありません")

            sql = "SELECT * FROM `transaction_evidences` WHERE `id` = ? FOR UPDATE"
            c.execute(sql, transaction_evidence["id"])
            transaction_evidence = c.fetchone()
            if transaction_evidence is None:
                conn.rollback()
                http_json_error(requests.codes["not_found"], "transaction_evidences not found")
            if transaction_evidence["status"] != Constants.TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING:
                conn.rollback()
                http_json_error(requests.codes['forbidden'], "準備ができていません")

            sql = "SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE"
            c.execute(sql, transaction_evidence["id"])
            shipping = c.fetchone()
            if shipping is None:
                conn.rollback()
                http_json_error(requests.codes["not_found"], "shipping not found")

            try:
                host = get_shipment_service_url()
                res = requests.post(host + "/request",
                                    header=dict(Authorization=Constants.ISUCARI_API_TOKEN),
                                    json=dict(reserve_id=shipping["reserve_id"]))
                res.raise_for_status()
            except (socket.gaierror, requests.HTTPError) as err:
                conn.rollback()
                app.logger.exception(err)
                http_json_error(requests.codes["internal_server_error"], "failed to request to shipment service")

            sql = "UPDATE `shippings` SET `status` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?"
            c.execute(sql, (
                Constants.TRANSACTION_EVIDENCE_STATUS_WAIT_DONE,
                datetime.datetime.now(),
                transaction_evidence["id"],
            ))
        conn.commit()
    except MySQLdb.Error as err:
        app.logger.exception(err)
        http_json_error(requests.codes['internal_server_error'], "db error")
    return flask.jsonify(dict(transaction_evidence_id=transaction_evidence["id"]))


@app.route("/complete", methods=["POST"])
def post_complete():
    ensure_valid_csrf_token()
    user = get_user()
    conn = dbh()
    with conn.cursor() as c:
        try:
            sql = "SELECT * FROM `transaction_evidences` WHERE `item_id` = ?"
            c.execute(sql, flask.request.json["item_id"])
            transaction_evidence = c.fetchone()
            if transaction_evidence is None:
                http_json_error(requests.codes["not_found"], "transaction_evidences not found")
        except MySQLdb.Error as err:
            app.logger.exception(err)
            http_json_error(requests.codes['internal_server_error'], "db error")
    if transaction_evidence["seller_id"] != user["id"]:
        http_json_error(requests.codes['forbidden'], "権限がありません")

    try:
        conn.begin()
        with conn.cursor() as c:
            sql = "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE"
            c.execute(sql, flask.request.json["item_id"])
            item = c.fetchone()
            if item is None:
                conn.rollback()
                http_json_error(requests.codes["not_found"], "item not found")
            if item["status"] != Constants.ITEM_STATUS_TRADING:
                conn.rollback()
                http_json_error(requests.codes["forbidden"], "商品が取引中ではありません")

            sql = "SELECT * FROM `transaction_evidences` WHERE `item_id` = ? FOR UPDATE"
            c.execute(sql, flask.request.json["item_id"])
            transaction_evidence = c.fetchone()
            if transaction_evidence is None:
                conn.rollback()
                http_json_error(requests.codes["not_found"], "transaction_evidences not found")
            if transaction_evidence["status"] != Constants.TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING:
                conn.rollback()
                http_json_error(requests.codes['forbidden'], "準備ができていません")

            sql = "SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE"
            c.execute(sql, transaction_evidence["id"])
            shipping = c.fetchone()

            try:
                host = get_shipment_service_url()
                res = requests.post(host + "/request",
                                    header=dict(Authorization=Constants.ISUCARI_API_TOKEN),
                                    json=dict(reserve_id=shipping["reserve_id"]))
                res.raise_for_status()
            except (socket.gaierror, requests.HTTPError) as err:
                conn.rollback()
                app.logger.exception(err)
                http_json_error(requests.codes["internal_server_error"], "failed to request to shipment service")

            if item["status"] != Constants.SHIPPING_STATUS_DONE:
                conn.rollback()
                http_json_error(requests.codes["bad_request"], "shipment service側で配送完了になっていません")

            sql = "UPDATE `shippings` SET `status` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?"
            c.execute(sql, (
                Constants.SHIPPING_STATUS_DONE,
                datetime.datetime.now(),
                transaction_evidence["id"],
            ))

            sql = "UPDATE `items` SET `status` = ?, `updated_at` = ? WHERE `id` = ?"
            c.execute(sql, (
                Constants.ITEM_STATUS_SOLD_OUT,
                datetime.datetime.now(),
                item["id"],
            ))

        conn.commit()
    except MySQLdb.Error as err:
        app.logger.exception(err)
        http_json_error(requests.codes['internal_server_error'], "db error")
    return flask.jsonify(dict(transaction_evidence_id=transaction_evidence["id"]))


@app.route("/transactions/<transaction_evidence_id>.png", methods=["GET"])
def get_qrcode():
    return


@app.route("/bump", methods=["POST"])
def post_bump():
    ensure_valid_csrf_token()
    ensure_required_payload(['item_id'])
    user = get_user()

    try:
        conn = dbh()
        conn.begin()
        with conn.cursor() as c:
            sql = "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE"
            c.execute(sql, (flask.request.json['item_id']))
            target_item = c.fetchone()
            if target_item is None:
                conn.rollback()
                http_json_error(requests.codes['not_found'], "item not found")
            if target_item['seller_id'] != user['id']:
                conn.rollback()
                http_json_error(requests.codes['forbidden'], "自分の商品以外は編集できません")

            sql = "SELECT * FROM `users` WHERE `id` = ? FOR UPDATE"
            c.execute(sql, (user['id']))
            seller = c.fetchone()
            if seller is None:
                conn.rollback()
                http_json_error(requests.codes['not_found'], "user not found")
            now = datetime.datetime.now()
            if seller['last_bump'] + datetime.timedelta(seconds=3) < now:
                http_json_error(requests.codes['forbidden'], "Bump not allowed")

            sql = "UPDATE `items` SET `created_at`=?, `updated_at`=? WHERE id=?"
            c.execute(sql, (target_item['id']))

            sql = "UPDATE `users` SET `last_bump`=? WHERE id=?"
            c.execute(sql, (now, user['id']))

            sql = "SELECT * FROM `items` WHERE `id` = ?"
            c.execute(sql, target_item['id'])
            target_item = c.fetchone()

        conn.commit()
    except MySQLdb.Error as err:
        app.logger.exception(err)
        http_json_error(requests.codes['internal_server_error'], "db error")

    return flask.jsonify({
        'item_id': target_item['id'],
        'item_price': target_item['price'],
        'item_created_at': int(target_item['created_at'].timestamp()),
        'item_updated_at': int(target_item['updated_at'].timestamp()),
    })


@app.route("/settings", methods=["GET"])
def get_settings():
    try:
        conn = dbh()
        sql = "SELECT * FROM `categories`"
        with conn.cursor() as c:
            c.execute(sql)
            categories = c.fetchall()
    except MySQLdb.Error as err:
        app.logger.exception(err)
        http_json_error(requests.codes['internal_server_error'], "db error")

    return flask.jsonify(dict(
        user=to_user_json(get_user()),
        csrf_token=flask.session.get('csrf_token'),
        categories=categories))


@app.route("/login", methods=["POST"])
def post_login():
    ensure_required_payload(['account_name', 'password'])
    try:
        conn = dbh()
        sql = "SELECT * FROM `users` WHERE `account_name` = %s"
        with conn.cursor() as c:
            c.execute(sql, [flask.request.json['account_name']])
            user = c.fetchone()

            if user is None or \
                    not bcrypt.checkpw(flask.request.json['password'].encode('utf-8'), user['hashed_password']):
                http_json_error(requests.codes['unauthorized'], 'アカウント名かパスワードが間違えています')
    except MySQLdb.Error as err:
        app.logger.exception(err)
        http_json_error(requests.codes['internal_server_error'], 'db error')

    flask.session['user_id'] = user['id']
    flask.session['csrf_token'] = random_string(10)
    return flask.jsonify(
        to_user_json(user),
    )


@app.route("/register", methods=["POST"])
def post_register():
    ensure_required_payload(['account_name', 'password', 'address'])
    hashedpw = bcrypt.hashpw(flask.request.json['password'].encode('utf-8'), bcrypt.gensalt(10))
    try:
        conn = dbh()
        with conn.cursor() as c:
            sql = "INSERT INTO `users` (`account_name`, `hashed_password`, `address`) VALUES (%s, %s, %s)"
            c.execute(sql, [flask.request.json['account_name'], hashedpw, flask.request.json['address']])
        conn.commit()
        user_id = c.lastrowid
    except MySQLdb.Error as err:
        app.logger.exception(err)
        http_json_error(requests.codes['internal_server_error'], 'db error')

    flask.session['user_id'] = user_id
    flask.session['csrf_token'] = random_string(10)
    return flask.jsonify({
        'id': user_id,
        'account_name': flask.request.json['account_name'],
        'address': flask.request.json['address'],
    })


@app.route("/reports.json", methods=["GET"])
def get_reports():
    try:
        conn = dbh()
        conn.begin()
        with conn.cursor() as c:
            sql = "SELECT * FROM `transaction_evidences` WHERE `id` > 15007"
            c.execute(sql)
            transaction_evidences = c.fetchall()
    except MySQLdb.Error as err:
        app.logger.exception(err)
        http_json_error(requests.codes['internal_server_error'], "db error")
    return flask.jsonify(dict(transaction_evidences=transaction_evidences))


# Frontend
@app.route("/")
@app.route("/login")
@app.route("/register")
@app.route("/timeline")
@app.route("/categories/<category_id>/items")
@app.route("/sell")
@app.route("/items/<item_id>")
@app.route("/items/<item_id>/edit")
@app.route("/items/<item_id>/buy")
@app.route("/buy/compelete")
@app.route("/transactions/<transaction_id>")
@app.route("/users/<user_id>")
@app.route("/users/setting")
def get_index():
    # if "user_id" in flask.session:
    #    return flask.redirect('/', 303)
    return flask.render_template('index.html')


## Assets
# @app.route("/*")

if __name__ == "__main__":
    app.run(port=8080, debug=True, threaded=True)
