#!/usr/bin/env python

import MySQLdb.cursors
import flask
import os
import json
import pathlib

base_path = pathlib.Path(__file__).resolve().parent.parent
static_folder = base_path / 'public'

app = flask.Flask(__name__, static_folder = str(static_folder), static_url_path = '')
app.config['SECRET_KEY'] = 'isucari'

def dbh():
    if hastttr(flask.g, 'db'):
        return flask.g.db

    flask.g.db = MySQLdb.connect(
        host = os.environ["MYSQL_HOST"],
        port = "3306",
        user = os.environ["MYSQL_USER"],
        password = os.environ["MYSQL_PASS"],
        dbname = os.environ["MYSQL_DBNAME"],
        charset = 'utf8mb4',
        cursorclass = MySQLdb.cursors.DictCursor,
        autocommit = True,
    )
    cur = flask.g.db.cursor()
    cur.execute("SET SESSION sql_mode='STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION'")
    return flask.g.db

# API
@app.route("/initialize", methods=["POST"])
def post_initialize():
        subprocess.call(["../sql/init.sh"])
        return ('', 204)

@app.route("/new_items.json", methods=["GET"])
def get_new_items():
    return

@app.route("/new_items/<root_category_id>.json", methods=["GET"])
def get_new_category_items(root_category_id = None):
    return

@app.route("/users/transactions.json", methods=["GET"])
def get_transactions():
    return

@app.route("/users/<user_id>.json", methods=["GET"])
def get_user_items(user_id = None):
    return

@app.route("/items/<item_id>.json", methods=["GET"])
def get_item(item_id = None):
    return

@app.route("/itemds/edit", methods=["POST"])
def post_item_edit():
    return

@app.route("/buy", methods=["POST"])
def post_buy():
    return

@app.route("/sell", methods=["POST"])
def post_sell():
    return

@app.route("/ship", methods=["POST"])
def post_ship():
    return

@app.route("/ship_done", methods=["POST"])
def post_ship_done():
    return

@app.route("/complete", methods=["POST"])
def post_complete():
    return

@app.route("/transactions/<transaction_evidence_id>.png", methods=["GET"])
def get_qrcode():
    return

@app.route("/bump", methods=["POST"])
def post_bump():
    return

@app.route("/settings", methods=["GET"])
def get_settings():
    return

@app.route("/login", methods=["POST"])
def post_login():
    return

@app.route("/register", methods=["POST"])
def post_register():
    return

@app.route("/reports.json", methods=["GET"])
def get_reports():
    return


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
    #if "user_id" in flask.session:
    #    return flask.redirect('/', 303)
    return flask.render_template('index.html')

## Assets
#@app.route("/*")

if __name__ == "__main__":
    app.run(port = 8080, debug = True, threaded = True)

