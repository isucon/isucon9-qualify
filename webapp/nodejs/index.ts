import {IncomingMessage, ServerResponse} from "http";
import util, {isNullOrUndefined, types} from "util";
import childProcess from "child_process";
import path from "path";
import fs from "fs";

import TraceError from "trace-error";
import createFastify, {FastifyRequest, FastifyReply} from "fastify";
// @ts-ignore
import fastifyMysql from "fastify-mysql";
import fastifyCookie from "fastify-cookie";
import fastifyStatic from "fastify-static";
import crypt from "crypto";
import bcrypt from "bcrypt";
import {create} from "domain";

const execFile = util.promisify(childProcess.execFile);

type MySQLResultRows = Array<any> & { insertId: number };
type MySQLColumnCatalogs = Array<any>;

type MySQLResultSet = [MySQLResultRows, MySQLColumnCatalogs];

interface MySQLQueryable {
    query(sql: string, params?: ReadonlyArray<any>): Promise<MySQLResultSet>;
}

interface MySQLClient extends MySQLQueryable {
    beginTransaction(): Promise<void>;

    commit(): Promise<void>;

    rollback(): Promise<void>;

    release(): void;
}

declare module "fastify" {
    interface FastifyInstance<HttpServer, HttpRequest, HttpResponse> {
        mysql: MySQLQueryable & {
            getConnection(): Promise<MySQLClient>;
        };
    }

    interface FastifyRequest<HttpRequest> {
        // add types if needed
    }

    interface FastifyReply<HttpResponse> {
        // add types if needed
    }
}

// =============================================

function TODO() {
    throw new Error("Not yet implemented!");
}

const DefaultPaymentServiceURL = "http://localhost:5555";
const DefaultShipmentServiceURL = "http://localhost:7000";
const ItemMinPrice = 100;
const ItemMaxPrice = 1000000;
const ItemPriceErrMsg =
    "商品価格は100ｲｽｺｲﾝ以上、1,000,000ｲｽｺｲﾝ以下にしてください";
const ItemStatusOnSale = "on_sale";
const ItemStatusTrading = "trading";
const ItemStatusSoldOut = "sold_out";
const ItemStatusStop = "stop";
const ItemStatusCancel = "cancel";
const PaymentServiceIsucariAPIKey = "a15400e46c83635eb181-946abb51ff26a868317c";
const PaymentServiceIsucariShopID = "11";
const TransactionEvidenceStatusWaitShipping = "wait_shipping";
const TransactionEvidenceStatusWaitDone = "wait_done";
const TransactionEvidenceStatusDone = "done";
const ShippingsStatusInitial = "initial";
const ShippingsStatusWaitPickup = "wait_pickup";
const ShippingsStatusShipping = "shipping";
const ShippingsStatusDone = "done";
const BumpChargeSeconds = 3;
const ItemsPerPage = 48;
const TransactionsPerPage = 10;
const BcryptCost = 10;

type Config = {
    name: string;
    val: string;
};

type User = {
    id: number;
    account_name: string;
    hashed_password: string;
    address: string;
    num_sell_items: number;
    last_bump: Date;
    created_at: Date;
};

type UserSimple = {
    id: number;
    account_name: string;
    num_sell_items: number;
};

type Item = {
    id: number;
    seller_id: number;
    buyer_id: number;
    status: string;
    name: string;
    price: number;
    description: string;
    image_name: string;
    category_id: number;
    created_at: Date;
    updated_at: Date;
};

type ItemSimple = {
    id: number;
    seller_id: number;
    seller: UserSimple;
    status: string;
    name: string;
    price: number;
    image_url: string;
    category_id: number;
    category: Category;
    created_at: number;
};

type ItemDetail = {
    id: number;
    seller_id: number;
    seller: UserSimple;
    buyer_id?: number;
    buyer?: UserSimple;
    status: string;
    name: string;
    price: number;
    description: string;
    image_url: string;
    category_id: number;
    category: Category;
    transaction_evidence_id?: number;
    transaction_evidence_status?: string;
    shipping_status?: string;
    created_at: number;
};

type TransactionEvidence = {
    id: number;
    seller_id: number;
    buyer_id: string;
    status: string;
    item_id: string;
    item_name: string;
    item_price: number;
    item_description: string;
    item_category_id: number;
    item_root_category_id: number;
    created_at: Date;
    updated_at: Date;
};

type Shipping = {
    transaction_evidence_id: number;
    status: string;
    item_name: string;
    item_id: number;
    reserve_id: string;
    reserve_time: number;
    to_address: string;
    to_name: string;
    from_address: string;
    from_name: string;
    img_binary: Uint8Array,
};

type Category = {
    id: number,
    parent_id: number,
    category_name: string,
    parent_category_name: string,
};

type ReqInitialize = {
    payment_service_url: string;
    shipment_service_url: string;
};

type ReqRegister = {
    account_name?: string,
    address?: string,
    password?: string,
}

type ReqLogin = {
    account_name?: string,
    password?: string,
}

type ResNewItems = {
    root_category_id?: number,
    root_category_name?: string,
    has_next: boolean,
    items: ItemSimple[],
}

type ResUserItems = {
    user: UserSimple,
    has_next: boolean,
    items: ItemSimple[],
}

const fastify = createFastify({
    logger: true
});

fastify.register(fastifyStatic, {
    root: path.join(__dirname, "public")
});

fastify.register(fastifyCookie);

fastify.register(fastifyMysql, {
    host: process.env.MYSQL_HOST || "127.0.0.1",
    port: process.env.MYSQL_PORT || "3306",
    user: process.env.MYSQL_USER || "isucari",
    password: process.env.MYSQL_PASS || "isucari",
    database: process.env.MYSQL_DBNAME || "isucari",

    promise: true
});

function buildUriFor<T extends IncomingMessage>(request: FastifyRequest<T>) {
    const uriBase = `http://${request.headers.host}`;
    return (path: string) => {
        return `${uriBase}${path}`;
    };
}

async function getConnection() {
    return fastify.mysql.getConnection();
}

// API

fastify.post("/initialize", postInitialize);
fastify.get("/new_items.json", getNewItems);
fastify.get("/new_items/:root_category_id(^\\d+).json", getNewCategoryItems);
fastify.get("/users/transactions.json", getTransactions);
fastify.get("/users/:user_id(^\\d+).json", getUserItems);
fastify.get("/items/:item_id(^\\d+).json", getItem);
fastify.post("/items/edit", postItemEdit);
fastify.post("/buy", postBuy);
fastify.post("/sell", postSell);
fastify.post("/ship", postShip)
fastify.post("/ship_done", postShipDone);
fastify.post("/complete", postComplete);
fastify.get("/transactions/:transaction_evidence_id(^\\d+).png", getQRCode);
fastify.post("/bump", postBump);
fastify.get("/settings", getSettings);
fastify.post("/login", postLogin);
fastify.post("/register", postRegister);
fastify.get("/reports.json", getReports);

// Frontend
fastify.get("/", getIndex);
fastify.get("/login", getIndex);
fastify.get("/register", getIndex);
fastify.get("/timeline", getIndex);
fastify.get("/categories/:category_id/items", getIndex);
fastify.get("/sell", getIndex);
fastify.get("/items/:item_id/edit", getIndex);
fastify.get("/items/:item_id/buy", getIndex);
fastify.get("/buy/complete", getIndex);

async function getIndex(_req: any, reply: FastifyReply<ServerResponse>) {
    const html = await fs.promises.readFile(
        path.join(__dirname, "public/index.html")
    );
    reply.type("text/html").send(html);
}

async function postInitialize(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const ri: ReqInitialize = req.body;

    await execFile("../sql/init.sh");

    const conn = await getConnection();

    await conn.query(
        "INSERT INTO `configs` (`name`, `val`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `val` = VALUES(`val`)",
        ["payment_service_url", ri.payment_service_url]
    );

    await conn.query(
        "INSERT INTO `configs` (`name`, `val`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `val` = VALUES(`val`)",
        ["shipment_service_url", ri.shipment_service_url]
    );

    const res = {
        // Campaign 実施時は true にする
        is_campaign: false
    };

    reply
        .code(200)
        .type("application/json")
        .send(res);
}

async function getNewItems(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const query = req.query;
    let itemId = 0;
    if (query['item_id'] !== undefined) {
        itemId = parseInt(query['item_id'], 10);
        if (isNaN(itemId) || itemId <= 0) {
            outputErrorMessage(reply, "item_id param error", 400);
            return
        }
    }

    let createdAt = 0;
    if (query['created_at'] !== undefined) {
        createdAt = parseInt(query['created_at'], 10);
        if (isNaN(createdAt) || createdAt <= 0) {
            outputErrorMessage(reply, "created_at param error", 400);
            return
        }
    }

    const items: Item[] = [];
    const conn = await getConnection();
    if (itemId > 0 && createdAt > 0) {
        const [rows,] = await conn.query(
            "SELECT * FROM `items` WHERE `status` IN (?,?) AND `created_at` <= ? AND `id` < ? ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
            [
                ItemStatusOnSale,
                ItemStatusSoldOut,
                new Date(createdAt),
                itemId,
                ItemsPerPage + 1,
            ],
        );
        for (const row of rows) {
            items.push(row as Item);
        }
    } else {
        const [rows,] = await conn.query(
            "SELECT * FROM `items` WHERE `status` IN (?,?) ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
            [
                ItemStatusOnSale,
                ItemStatusSoldOut,
                ItemsPerPage + 1,
            ],
        );
        for (const row of rows) {
            items.push(row as Item);
        }
    }

    let itemSimples: ItemSimple[] = [];

    for (const item of items) {
        const seller = await getUserSimpleByID(conn, item.seller_id);
        if (seller === null) {
            outputErrorMessage(reply, "seller not found", 404)
            return;
        }
        const category = await getCategoryByID(conn, item.category_id);
        if (category === null) {
            outputErrorMessage(reply, "category not found", 404)
            return;
        }

        itemSimples.push({
            id: item.id,
            seller_id: item.seller_id,
            seller: seller,
            status: item.status,
            name: item.name,
            price: item.price,
            image_url: getImageURL(item.image_name),
            category_id: item.category_id,
            category: category,
            created_at: item.created_at.getTime(),
        });
    }

    let hasNext = false;
    if (itemSimples.length > ItemsPerPage) {
        hasNext = true;
        itemSimples = itemSimples.splice(0, itemSimples.length - 1)
    }
    const res: ResNewItems = {
        has_next: hasNext,
        items: itemSimples,
    };

    reply
        .code(200)
        .type("application/json")
        .send(res);
}

async function getNewCategoryItems(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const rootCategoryIdStr: string = req.params.root_category_id;
    const rootCategoryId: number = parseInt(rootCategoryIdStr, 10);
    if (rootCategoryId === null || isNaN(rootCategoryId)) {
        outputErrorMessage(reply, "incorrect category id", 400);
        return;
    }

    const conn = await getConnection();
    const rootCategory = await getCategoryByID(conn, rootCategoryId);
    if (rootCategory === null || rootCategory.parent_id !== 0) {
        outputErrorMessage(reply, "category not found");
        return;
    }

    const categoryIDs: number[] = [];
    const [rows,] = await conn.query("SELECT id FROM `categories` WHERE parent_id=?", [rootCategory.id]);
    for (const row of rows) {
        categoryIDs.push(row.id);
    }

    const itemIDStr = req.query.item_id;
    let itemID = 0;
    if (itemIDStr !== undefined && itemIDStr !== "") {
        itemID = parseInt(itemIDStr, 10);
        if (isNaN(itemID) || itemID <= 0) {
            outputErrorMessage(reply, "item_id param error", 400);
            return;
        }
    }
    const createdAtStr = req.query.created_at;
    let createdAt = 0;
    if (createdAtStr !== undefined && createdAtStr !== "") {
        createdAt = parseInt(createdAtStr, 10);
        if (isNaN(createdAt) || createdAt <= 0) {
            outputErrorMessage(reply, "created_at param error", 400);
            return;
        }
    }

    const items: Item[] = [];
    if (itemID > 0 && createdAt > 0) {
        const [rows] = await conn.query(
            "SELECT * FROM `items` WHERE `status` IN (?,?) AND `created_at` <= ? AND `id` < ? ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
            [
                ItemStatusOnSale,
                ItemStatusSoldOut,
                new Date(createdAt),
                itemID,
                ItemsPerPage + 1,
            ]
        );

        for (const row of rows) {
            items.push(row as Item);
        }
    } else {
        const [rows] = await conn.query(
            "SELECT * FROM `items` WHERE `status` IN (?,?) ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
            [
                ItemStatusOnSale,
                ItemStatusSoldOut,
                ItemsPerPage + 1,
            ]
        );

        for (const row of rows) {
            items.push(row as Item);
        }
    }

    let itemSimples: ItemSimple[] = [];

    for (const item of items) {
        const seller = await getUserSimpleByID(conn, item.seller_id);
        if (seller === null) {
            outputErrorMessage(reply, "seller not found", 404)
            return;
        }
        const category = await getCategoryByID(conn, item.category_id);
        if (category === null) {
            outputErrorMessage(reply, "category not found", 404)
            return;
        }

        itemSimples.push({
            id: item.id,
            seller_id: item.seller_id,
            seller: seller,
            status: item.status,
            name: item.name,
            price: item.price,
            image_url: getImageURL(item.image_name),
            category_id: item.category_id,
            category: category,
            created_at: item.created_at.getTime(),
        });
    }

    let hasNext = false;
    if (itemSimples.length > ItemsPerPage) {
        hasNext = true;
        itemSimples = itemSimples.splice(0, itemSimples.length - 1)
    }

    const res = {
        root_category_id: rootCategory.id,
        root_category_name: rootCategory.category_name,
        items: itemSimples,
        has_next: hasNext,
    }

    reply
        .code(200)
        .type("application/json")
        .send(res);

}

async function getTransactions(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const conn = await getConnection();
    const user = await getLoginUser(req, conn);

    if (user === null) {
        outputErrorMessage(reply, "no session", 404);
        return;
    }

    const query = req.query;
    let itemId = 0;
    if (query['item_id'] !== undefined) {
        itemId = parseInt(query['item_id'], 10);
        if (isNaN(itemId) || itemId <= 0) {
            outputErrorMessage(reply, "item_id param error", 400);
            return
        }
    }

    let createdAt = 0;
    if (query['created_at'] !== undefined) {
        createdAt = parseInt(query['created_at'], 10);
        if (isNaN(createdAt) || createdAt <= 0) {
            outputErrorMessage(reply, "created_at param error", 400);
            return
        }
    }

    await conn.beginTransaction();
    const items: Item[] = [];
    if (itemId > 0 && createdAt > 0) {
        const [rows] = await conn.query(
            "SELECT * FROM `items` WHERE (`seller_id` = ? OR `buyer_id` = ?) AND `status` IN (?,?,?,?,?) AND `created_at` <= ? AND `id` < ? ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
            [
                user.id,
                user.id,
                ItemStatusOnSale,
                ItemStatusTrading,
                ItemStatusSoldOut,
                ItemStatusCancel,
                ItemStatusStop,
                new Date(createdAt),
                itemId,
                TransactionsPerPage + 1,
            ]
        );

        for (const row of rows) {
            items.push(row as Item);
        }

    } else {
        const [rows] = await conn.query(
            "SELECT * FROM `items` WHERE (`seller_id` = ? OR `buyer_id` = ?) AND `status` IN (?,?,?,?,?) ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
            [
                user.id,
                user.id,
                ItemStatusOnSale,
                ItemStatusTrading,
                ItemStatusSoldOut,
                ItemStatusCancel,
                ItemStatusStop,
                TransactionsPerPage + 1
            ]
        );

        for (const row of rows) {
            items.push(row as Item);
        }
    }

    let itemDetails: ItemDetail[] = [];
    for (const item of items) {
        const category = await getCategoryByID(conn, item.category_id);
        if (category === null) {
            outputErrorMessage(reply, "category not found", 404)
            await conn.rollback();
            return;
        }

        const seller = await getUserSimpleByID(conn, item.seller_id);
        if (seller === null) {
            outputErrorMessage(reply, "seller not found", 404)
            await conn.rollback();
            return;
        }

        const itemDetail: ItemDetail = {
            id: item.id,
            seller_id: item.seller_id,
            seller: seller,
            // buyer_id
            // buyer
            status: item.status,
            name: item.name,
            price: item.price,
            description: item.description,
            image_url: getImageURL(item.image_name),
            category_id: item.category_id,
            category: category,
            // transaction_evidence_id
            // transaction_evidence_status
            // shipping_status
            created_at: item.created_at.getTime(),
        };

        if (item.buyer_id !== undefined) {
            const buyer = await getUserSimpleByID(conn, item.buyer_id);
            if (buyer === null) {
                outputErrorMessage(reply, "buyer not found", 404);
                await conn.rollback();
                return;
            }
        }

        const [rows] = await conn.query("SELECT * FROM `transaction_evidences` WHERE `item_id` = ?", [item.id]);
        let transactionEvidence: TransactionEvidence | null = null;
        for (const row of rows) {
            transactionEvidence = row as TransactionEvidence;
        }

        if (transactionEvidence !== null) {
            const [rows] = await conn.query("SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ?", [transactionEvidence.id]);

            let shipping: Shipping | null = null;
            for (const row of rows) {
                shipping = row as Shipping;
            }

            if (shipping === null) {
                outputErrorMessage(reply, "shipping not found", 404);
                await conn.rollback();
                return;
            }

            // TODO APIShipmentStatus

            itemDetail.transaction_evidence_id = transactionEvidence.id;
            itemDetail.transaction_evidence_status = transactionEvidence.status;
            itemDetail.shipping_status = ShippingsStatusDone; // TODO
        }

        itemDetails.push(itemDetail);

    }

    await conn.commit();

    let hasNext = false;
    if (itemDetails.length > TransactionsPerPage) {
        hasNext = true;
        itemDetails = itemDetails.slice(0, TransactionsPerPage);
    }

    reply
        .code(200)
        .type("application/json;charset=utf-8")
        .send({has_next: hasNext, items: itemDetails});

}

async function getUserItems(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const userIdStr = req.params.user_id;
    const userId = parseInt(userIdStr, 10);
    if (userId === undefined || isNaN(userId)) {
        outputErrorMessage(reply, "incorrect user id", 400);
        return;
    }

    const conn = await getConnection();
    const userSimple = await getUserSimpleByID(conn, userId);
    if (userSimple === null) {
        outputErrorMessage(reply, "user not found", 404);
        return;
    }

    const itemIDStr = req.query.item_id;
    let itemID = 0;
    if (itemIDStr !== undefined && itemIDStr !== "") {
        itemID = parseInt(itemIDStr, 10);
        if (isNaN(itemID) || itemID <= 0) {
            outputErrorMessage(reply, "item_id param error", 400);
            return;
        }
    }
    const createdAtStr = req.query.created_at;
    let createdAt = 0;
    if (createdAtStr !== undefined && createdAtStr !== "") {
        createdAt = parseInt(createdAtStr, 10);
        if (isNaN(createdAt) || createdAt <= 0) {
            outputErrorMessage(reply, "created_at param error", 400);
            return;
        }
    }

    const items: Item[] = [];
    if (itemID > 0 && createdAt > 0) {
        const [rows] = await conn.query(
            "SELECT * FROM `items` WHERE `seller_id` = ? AND `status` IN (?,?,?) AND `created_at` <= ? AND `id` < ? ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
            [
                userSimple.id,
                ItemStatusOnSale,
                ItemStatusTrading,
                ItemStatusSoldOut,
                new Date(createdAt),
                itemID,
                ItemsPerPage + 1,
            ]
        );

        for (const row of rows) {
            items.push(row as Item);
        }
    } else {
        const [rows] = await conn.query(
            "SELECT * FROM `items` WHERE `seller_id` = ? AND `status` IN (?,?,?) ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
            [
                userSimple.id,
                ItemStatusOnSale,
                ItemStatusTrading,
                ItemStatusSoldOut,
                ItemsPerPage + 1,
            ]
        );

        for (const row of rows) {
            items.push(row as Item);
        }
    }

    let itemSimples: ItemSimple[] = [];
    for (const item of items) {
        const category = await getCategoryByID(conn, item.category_id);
        if (category === null) {
            outputErrorMessage(reply, "category not found", 404)
            return;
        }

        itemSimples.push({
            id: item.id,
            seller_id: item.seller_id,
            seller: userSimple,
            status: item.status,
            name: item.name,
            price: item.price,
            image_url: getImageURL(item.image_name),
            category_id: item.category_id,
            category: category,
            created_at: item.created_at.getTime(),
        });
    }

    let hasNext = false;
    if (itemSimples.length > ItemsPerPage) {
        hasNext = true;
        itemSimples = itemSimples.slice(0, ItemsPerPage);
    }
    const res: ResUserItems = {
        user: userSimple,
        has_next: hasNext,
        items: itemSimples,
    };

    reply
        .code(200)
        .type("application/json")
        .send(res);
}

async function getItem(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const itemIdStr = req.params.item_id;
    const itemId = parseInt(itemIdStr, 10);
    if (itemId === undefined || isNaN(itemId)) {
        outputErrorMessage(reply, "incorrect item id", 400);
        return;
    }

    const conn = await getConnection();
    const user = await getLoginUser(req, conn);
    if (user === null) {
        outputErrorMessage(reply, "no session", 404);
        return;
    }

    const [rows] = await conn.query("SELECT * FROM `items` WHERE `id` = ?", [itemId]);
    let item: Item | null = null;

    for (const row of rows) {
        item = row as Item;
    }

    if (item === null) {
        outputErrorMessage(reply, "item not found", 404);
        return;
    }

    const category = await getCategoryByID(conn, item.category_id);
    if (category === null) {
        outputErrorMessage(reply, "category not found", 404)
        return;
    }

    const seller = await getUserSimpleByID(conn, item.seller_id);
    if (seller === null) {
        outputErrorMessage(reply, "seller not found", 404)
        return;
    }

    const itemDetail: ItemDetail = {
        id: item.id,
        seller_id: item.seller_id,
        seller: seller,
        // buyer_id
        // buyer
        status: item.status,
        name: item.name,
        price: item.price,
        description: item.description,
        image_url: getImageURL(item.image_name),
        category_id: item.category_id,
        category: category,
        // transaction_evidence_id
        // transaction_evidence_status
        // shipping_status
        created_at: item.created_at.getTime(),
    };

    if ((user.id === item.seller_id || user.id === item.buyer_id) && item.buyer_id === undefined) {
        const buyer = await getUserSimpleByID(conn, item.buyer_id);
        if (buyer === null) {
            outputErrorMessage(reply, "buyer not found", 404);
            return;
        }

        itemDetail.buyer_id = item.buyer_id;
        itemDetail.buyer = buyer;

        const [rows] = await conn.query("SELECT * FROM `transaction_evidences` WHERE `item_id` = ?", [item.id]);
        let transactionEvidence: TransactionEvidence | null = null;
        for (const row of rows) {
            transactionEvidence = row as TransactionEvidence;
        }

        if (transactionEvidence !== null) {
            const [rows] = await conn.query("SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ?", [transactionEvidence.id])
            let shipping: Shipping | null = null;
            for (const row of rows) {
                shipping = row as Shipping;
            }

            if (shipping === null) {
                outputErrorMessage(reply, "shipping not found", 404);
                return;
            }

            itemDetail.transaction_evidence_id = transactionEvidence.id;
            itemDetail.transaction_evidence_status = transactionEvidence.status;
            itemDetail.shipping_status = shipping.status;
        }

    }

    reply
        .code(200)
        .type("application/json")
        .send(itemDetail);
}

async function postItemEdit(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const csrfToken = req.body.csrf_token;
    const itemID = req.body.item_id;
    const price = req.body.item_price;

    if (csrfToken !== req.cookies.csrf_token) {
        outputErrorMessage(reply, "csrf token error", 422)
        return;
    }

    if (price < ItemMinPrice || price > ItemMaxPrice) {
        outputErrorMessage(reply, ItemPriceErrMsg, 400);
        return;
    }

    const conn = await getConnection();

    const seller = await getLoginUser(req, conn);
    if (seller === null) {
        outputErrorMessage(reply, "no session", 404);
        return;
    }

    let targetItem: Item | null = null;
    ;
    {
        const [rows] = await conn.query("SELECT * FROM `items` WHERE `id` = ?", [itemID]);
        for (const row of rows) {
            targetItem = row as Item;
        }
    }

    if (targetItem === null) {
        outputErrorMessage(reply, "item not found");
        return;
    }

    if (targetItem.seller_id !== seller.id) {
        outputErrorMessage(reply, "自分の商品以外は編集できません", 403);
        return;
    }

    await conn.beginTransaction();

    await conn.query("SELECT * FROM `items` WHERE `id` = ? FOR UPDATE", [targetItem.id]);

    if (targetItem.status !== ItemStatusOnSale) {
        outputErrorMessage(reply, "販売中の商品以外編集できません", 403);
        await conn.rollback();
        return;
    }

    await conn.query("UPDATE `items` SET `price` = ?, `updated_at` = ? WHERE `id` = ?", [price, new Date(), targetItem.id]);

    {
        const [rows] = await conn.query("SELECT * FROM `items` WHERE `id` = ?", [targetItem.id]);
        for (const row of rows) {
            targetItem = row as Item;
        }
    }

    await conn.commit();

    reply
        .code(200)
        .type("application/json;charset=utf-8")
        .send({
            item_id: targetItem.id,
            item_price: targetItem.price,
            item_created_at: targetItem.created_at.getTime(),
            item_updated_at: targetItem.updated_at.getTime(),
        })


}

async function postBuy(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const csrfToken = req.body.csrf_token;

    if (csrfToken !== req.cookies.csrf_token) {
        outputErrorMessage(reply, "csrf token error", 422);
        return;
    }

    const conn = await getConnection();

    const buyer = await getLoginUser(req, conn);

    if (buyer === null) {
        outputErrorMessage(reply, "no session", 404);
        return;
    }

    await conn.beginTransaction();

    let targetItem: Item | null = null;
    {
        const [rows] = await conn.query("SELECT * FROM `items` WHERE `id` = ? FOR UPDATE", [req.body.item_id]);

        for (const row of rows) {
            targetItem = row as Item;
        }
    }

    if (targetItem === null) {
        outputErrorMessage(reply, "item not found", 404);
        await conn.rollback();
        return;
    }

    if (targetItem.status !== ItemStatusOnSale) {
        outputErrorMessage(reply, "item is not for sale", 403);
        await conn.rollback();
        return;
    }

    if (targetItem.seller_id === buyer.id) {
        outputErrorMessage(reply, "自分の商品は買えません", 403);
        await conn.rollback();
        return;
    }

    let seller: User | null = null;
    {
        const [rows] = await conn.query("SELECT * FROM `users` WHERE `id` = ? FOR UPDATE", [targetItem.seller_id]);
        for (const row of rows) {
            seller = row as User;
        }
    }

    if (seller === null) {
        outputErrorMessage(reply, "seller not found", 404);
        await conn.rollback();
        return;
    }

    const category = await getCategoryByID(conn, targetItem.category_id);
    if (category === null) {
        outputErrorMessage(reply, "category id error", 500);
        await conn.rollback();
        return;
    }

    const [result] = await conn.query(
        "INSERT INTO `transaction_evidences` (`seller_id`, `buyer_id`, `status`, `item_id`, `item_name`, `item_price`, `item_description`,`item_category_id`,`item_root_category_id`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
        [
            targetItem.seller_id,
            buyer.id,
            TransactionEvidenceStatusWaitShipping,
            targetItem.id,
            targetItem.name,
            targetItem.price,
            targetItem.description,
            category.id,
            category.parent_id,
        ]
    );

    const transactionEvidenceId = result.insertId;

    await conn.query(
        "UPDATE `items` SET `buyer_id` = ?, `status` = ?, `updated_at` = ? WHERE `id` = ?",
        [
            buyer.id,
            ItemStatusTrading,
            new Date(),
            targetItem.id,
        ]
    )

    // TODO ApiShipment.Create

    await conn.query(
        "INSERT INTO `shippings` (`transaction_evidence_id`, `status`, `item_name`, `item_id`, `reserve_id`, `reserve_time`, `to_address`, `to_name`, `from_address`, `from_name`, `img_binary`) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
        [
            transactionEvidenceId,
            ShippingsStatusInitial,
            targetItem.name,
            targetItem.id,
            "", // scr.reserve_id,
            "", // scr.reserve_time,
            buyer.address,
            buyer.account_name,
            seller.address,
            seller.account_name,
        ]
    );

    await conn.commit();

    reply.code(200)
        .type("application/json;charset=utf-8")
        .send({
            transaction_evidence_id: transactionEvidenceId,
        });

}

async function postSell(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
}

async function postShip(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
}

async function postShipDone(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
}

async function postComplete(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
}

async function getQRCode(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const transactionEvidenceIdStr: string = req.params.transaction_evidence_id;
    const transactionEvidenceId: number = parseInt(transactionEvidenceIdStr, 10);
    if (transactionEvidenceId === null || isNaN(transactionEvidenceId)) {
        outputErrorMessage(reply, "incorrect transaction_evidence id", 400);
        return;
    }

    const conn = await getConnection();
    const seller = await getLoginUser(req, conn);
    if (seller === null) {
        outputErrorMessage(reply, "no session", 404);
        return;
    }

    let transactionEvidence: TransactionEvidence | null = null;
    {
        const [rows] = await conn.query("SELECT * FROM `transaction_evidences` WHERE `id` = ?", [transactionEvidenceId]);
        for (const row of rows) {
            transactionEvidence = row as TransactionEvidence;
        }
    }

    if (transactionEvidence === null) {
        outputErrorMessage(reply, "transaction_evidence not found", 404);
        return;
    }

    if (transactionEvidence.seller_id !== seller.id) {
        outputErrorMessage(reply, "権限がありません", 403);
        return;
    }

    let shipping: Shipping | null = null;
    {
        const [rows] = await conn.query("SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ?", [transactionEvidence.id]);
        for (const row of rows) {
            shipping = row as Shipping;
        }
    }

    if (shipping === null) {
        outputErrorMessage(reply, "shippings not found", 404);
        return;
    }

    if (shipping.status !== ShippingsStatusWaitPickup && shipping.status !== ShippingsStatusShipping) {
        outputErrorMessage(reply, "qrcode not available", 403);
        return;
    }

    if (shipping.img_binary.byteLength === 0) {
        outputErrorMessage(reply, "empty qrcode image")
        return;
    }

    reply
        .code(200)
        .type("image/png")
        .send(shipping.img_binary);

}

async function postBump(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
}

async function getSettings(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const csrfToken = getCsrfToken(req);

    const conn = await getConnection();
    const user = await getLoginUser(req, conn);

    const res = {
        user: null as User | null,
        payment_service_url: null as string | null,
        categories: null as Category[] | null,
        csrf_token: null as string | null,
    };

    res.user = user;
    res.payment_service_url = getPaymentServiceURL();
    res.csrf_token = csrfToken;

    const categories: Category[] = [];
    const [rows] = await conn.query("SELECT * FROM `categories`", []);
    for (const row of rows) {
        categories.push(row as Category);
    }
    res.categories = categories;

    reply
        .code(200)
        .type("application/json")
        .send(res)

}

async function postLogin(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const rr: ReqLogin = req.body

    const accountName = rr.account_name;
    const password = rr.password;

    if (accountName === undefined || accountName === "" || password === undefined || password === "") {

        outputErrorMessage(reply, "all parameters are required", 400);
        return;
    }

    const conn = await getConnection();
    const [rows] = await conn.query("SELECT * FROM `users` WHERE `account_name` = ?", [accountName])
    let user: User | null = null;
    for (const row of rows) {
        user = row as User;
    }

    if (user === null) {
        outputErrorMessage(reply, "アカウント名かパスワードが間違えています", 401);
        return;
    }

    if (!await comparePassword(password, user.hashed_password)) {
        outputErrorMessage(reply, "アカウント名かパスワードが間違えています", 401);
        return;
    }

    reply.setCookie("user_id", user.id.toString(), {
        path: "/",
    });
    reply.setCookie("csrf_token", await getRandomString(128), {
        path: "/",
    });

    reply
        .code(200)
        .type("application/json;charset=utf-8")
        .send(user);

}

async function postRegister(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const rr: ReqRegister = req.body


    const accountName = rr.account_name;
    const address = rr.address;
    const password = rr.password;

    if (accountName === undefined || accountName === "" || password === undefined || password === "" || address === undefined || address === "") {
        outputErrorMessage(reply, "all parameters are required", 400);
        return;
    }

    const conn = await getConnection();

    const [rows] = await conn.query(
        "SELECT * FROM `users` WHERE `account_name` = ?",
        [
            accountName,
        ]
    );

    if (rows.length > 0) {
        outputErrorMessage(reply, "アカウント名かパスワードが間違えています", 401);
        return;
    }

    const hashedPassword = await encryptPassword(password);

    const [result,] = await conn.query(
        "INSERT INTO `users` (`account_name`, `hashed_password`, `address`) VALUES (?, ?, ?)",
        [
            accountName,
            hashedPassword,
            address,
        ]
    );

    const user = {
        id: result.insertId,
        account_name: accountName,
        address: address,
    };

    reply.setCookie("user_id", user.id.toString(), {
        path: "/",
    });

    reply.setCookie("csrf_token", await getRandomString(128), {
        path: "/",
    });

    reply
        .code(200)
        .type("application/json")
        .send(user);

}

async function getReports(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const conn = await getConnection();
    const [rows] = await conn.query("SELECT * FROM `transaction_evidences` WHERE `id` > 15007");
    const transactionEvidences: TransactionEvidence[] = [];
    for (const row of rows) {
        transactionEvidences.push(row as TransactionEvidence);
    }

    reply
        .code(200)
        .type("application/json")
        .send(transactionEvidences);
}

function getCsrfToken(req: FastifyRequest): string {
    return ""
}

async function getLoginUser(req: FastifyRequest, conn: MySQLQueryable): Promise<User | null> {
    let userId: number;
    if (req.cookies.user_id !== undefined && req.cookies.user_id !== "") {
        const [rows] = await conn.query("SELECT * FROM `users` WHERE `id` = ?", [req.cookies.user_id]);
        for (const row of rows) {
            const user = row as User;
            return user;
        }
    }

    return null;
}

function getPaymentServiceURL(): string {
    return "";
}

function getSession(req: FastifyRequest) {
}

fastify.listen(8000, (err, _address) => {
    if (err) {
        throw new TraceError("Failed to listening", err);
    }
});

function outputErrorMessage(reply: FastifyReply<ServerResponse>, message: string, status = 500) {
    reply.code(status)
        .type("application/json")
        .send({"error": message});
}

async function getUserSimpleByID(conn: MySQLQueryable, userID: number): Promise<UserSimple | null> {
    const [rows,] = await conn.query("SELECT * FROM `users` WHERE `id` = ?", [userID]);
    for (const row of rows) {
        const user = row as User;
        const userSimple: UserSimple = {
            id: user.id,
            account_name: user.account_name,
            num_sell_items: user.num_sell_items,
        };
        return userSimple;
    }
    return null;
}

async function getCategoryByID(conn: MySQLQueryable, categoryId: number): Promise<Category | null> {
    const [rows,] = await conn.query("SELECT * FROM `categories` WHERE `id` = ?", [categoryId]);
    for (const row of rows) {
        const category = row as Category;
        if (category.parent_id !== undefined && category.parent_id != 0) {
            const parentCategory = await getCategoryByID(conn, category.parent_id);
            if (parentCategory !== null) {
                category.parent_category_name = parentCategory.category_name
            }
        }
        return category;
    }
    return null;
}

async function encryptPassword(password: string): Promise<string> {
    return await new Promise((resolve) => {
        bcrypt.hash(password, 10, (err, hash) => {
            if (err != null) {
                throw err;
            }
            resolve(hash);
        });
    })
}

async function comparePassword(inputPassword: string, hashedPassword: string): Promise<boolean> {
    return await new Promise((resolve) => {
        bcrypt.compare(inputPassword, hashedPassword.toString(), (err, isValid) => {
            resolve(isValid);
        });
    });
}

async function getRandomString(length: number): Promise<string> {
    return await new Promise((resolve) => {
        crypt.randomBytes(length, (err, buffer) => {
            resolve(buffer.toString('hex'));
        })

    });
}

function getImageURL(image_name: string) {
    return "";
}

