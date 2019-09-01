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
import fastifySession from 'fastify-session'
import fastifyStatic from "fastify-static";

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

const sessionName = "session_isucari";
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
    buyer_id: number;
    buyer: UserSimple;
    status: string;
    name: string;
    price: number;
    description: string;
    image_url: string;
    category_id: number;
    category: Category;
    transaction_evidence_id: number;
    transaction_evidence_status: string;
    shipping_status: string;
    created_at: Date;
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

type Shipping = {};

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
fastify.register(fastifySession, {
    secret: '123456789012345678901234567890123',
    cookieName: sessionName,
    cookie: {secure: false}
});

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
fastify.get("/transactions/:transaction_evidence_id.png", getQRCode);
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
    console.log(req.params);
    const rootCategoryId: number = parseInt(rootCategoryIdStr, 10);
    if (rootCategoryId === null || isNaN(rootCategoryId)) {
        console.log(rootCategoryIdStr);
        console.log(rootCategoryId);
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
        itemSimples = itemSimples.splice(0, itemSimples.length - 1)
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

}

async function postItemEdit(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
}

async function postBuy(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
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
}

async function postBump(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
}

async function getSettings(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const csrfToken = getCsrfToken(req);

    const res = {
        User: null as User | null,
        PaymentServiceURL: null as string | null,
        Categories: null as Category[] | null
    };
    const user = await getUser(req);

    res.User = user;
    res.PaymentServiceURL = getPaymentServiceURL();

    const categories: Category[] = [];
    // TODO:
    res.Categories = categories;

    reply
        .code(200)
        .type("application/json")
        .send(res)

}

async function postLogin(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
}

async function postRegister(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
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

function getCsrfToken(req: FastifyRequest) {

}

async function getUser(req: FastifyRequest): Promise<User | null> {
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

function getImageURL(image_name: string) {
    return "";
}

