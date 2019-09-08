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
import fastifyMultipart from 'fastify-multipart';
import crypt from "crypto";
import bcrypt from "bcrypt";
import {paymentToken, shipmentCreate, shipmentRequest, shipmentStatus} from "./api";

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
const PaymentServiceIsucariShopID = 11;
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
    buyer_id: number;
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
    parent_category_name?: string,
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
    logger: {level: 'warn'}
});

fastify.register(fastifyStatic, {
    root: path.join(__dirname, "public")
});

fastify.register(fastifyCookie);

fastify.register(fastifyMultipart, {
    addToBody: true,
});

fastify.register(fastifyMysql, {
    host: process.env.MYSQL_HOST || "127.0.0.1",
    port: process.env.MYSQL_PORT || "3306",
    user: process.env.MYSQL_USER || "isucari",
    password: process.env.MYSQL_PASS || "isucari",
    database: process.env.MYSQL_DBNAME || "isucari",
    pool: 100,

    promise: true
});

function buildUriFor<T extends IncomingMessage>(request: FastifyRequest<T>) {
    const uriBase = `http://${request.headers.host}`;
    return (path: string) => {
        return `${uriBase}${path}`;
    };
}

async function getDBConnection() {
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
fastify.get("/items/:item_id", getIndex);
fastify.get("/items/:item_id/edit", getIndex);
fastify.get("/items/:item_id/buy", getIndex);
fastify.get("/buy/complete", getIndex);
fastify.get("/transactions/:transaction_id", getIndex);
fastify.get("/users/:user_id", getIndex);
fastify.get("/users/setting", getIndex);

async function getIndex(_req: any, reply: FastifyReply<ServerResponse>) {
    const html = await fs.promises.readFile(
        path.join(__dirname, "public/index.html")
    );
    reply.type("text/html").send(html);
}

async function postInitialize(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const ri: ReqInitialize = req.body;

    await execFile("../sql/init.sh");

    const db = await getDBConnection();

    await db.query(
        "INSERT INTO `configs` (`name`, `val`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `val` = VALUES(`val`)",
        ["payment_service_url", ri.payment_service_url]
    );

    await db.query(
        "INSERT INTO `configs` (`name`, `val`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `val` = VALUES(`val`)",
        ["shipment_service_url", ri.shipment_service_url]
    );

    const res = {
        // キャンペーン実施時には還元率の設定を返す。詳しくはマニュアルを参照のこと。
        campaign: 0,
        // 実装言語を返す
        language: "nodejs",
    };

    await db.release();

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
            replyError(reply, "item_id param error", 400);
            return
        }
    }

    let createdAt = 0;
    if (query['created_at'] !== undefined) {
        createdAt = parseInt(query['created_at'], 10);
        if (isNaN(createdAt) || createdAt <= 0) {
            replyError(reply, "created_at param error", 400);
            return
        }
    }

    const items: Item[] = [];
    const db = await getDBConnection();
    if (itemId > 0 && createdAt > 0) {
        const [rows,] = await db.query(
            "SELECT * FROM `items` WHERE `status` IN (?,?) AND (`created_at` < ? OR (`created_at` <= ? AND `id` < ?)) ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
            [
                ItemStatusOnSale,
                ItemStatusSoldOut,
                new Date(createdAt),
                new Date(createdAt),
                itemId,
                ItemsPerPage + 1,
            ],
        );
        for (const row of rows) {
            items.push(row as Item);
        }
    } else {
        const [rows,] = await db.query(
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
        const seller = await getUserSimpleByID(db, item.seller_id);
        if (seller === null) {
            replyError(reply, "seller not found", 404)
            await db.release();
            return;
        }
        const category = await getCategoryByID(db, item.category_id);
        if (category === null) {
            replyError(reply, "category not found", 404)
            await db.release();
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
        itemSimples = itemSimples.slice(0, itemSimples.length - 1)
    }
    const res: ResNewItems = {
        has_next: hasNext,
        items: itemSimples,
    };

    await db.release();

    reply
        .code(200)
        .type("application/json")
        .send(res);
}

async function getNewCategoryItems(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const rootCategoryIdStr: string = req.params.root_category_id;
    const rootCategoryId: number = parseInt(rootCategoryIdStr, 10);
    if (rootCategoryId === null || isNaN(rootCategoryId)) {
        replyError(reply, "incorrect category id", 400);
        return;
    }

    const db = await getDBConnection();
    const rootCategory = await getCategoryByID(db, rootCategoryId);
    if (rootCategory === null || rootCategory.parent_id !== 0) {
        replyError(reply, "category not found");
        await db.release();
        return;
    }

    const categoryIDs: number[] = [];
    const [rows,] = await db.query("SELECT id FROM `categories` WHERE parent_id=?", [rootCategory.id]);
    for (const row of rows) {
        categoryIDs.push(row.id);
    }

    const itemIDStr = req.query.item_id;
    let itemID = 0;
    if (itemIDStr !== undefined && itemIDStr !== "") {
        itemID = parseInt(itemIDStr, 10);
        if (isNaN(itemID) || itemID <= 0) {
            replyError(reply, "item_id param error", 400);
            await db.release();
            return;
        }
    }
    const createdAtStr = req.query.created_at;
    let createdAt = 0;
    if (createdAtStr !== undefined && createdAtStr !== "") {
        createdAt = parseInt(createdAtStr, 10);
        if (isNaN(createdAt) || createdAt <= 0) {
            replyError(reply, "created_at param error", 400);
            await db.release();
            return;
        }
    }

    const items: Item[] = [];
    if (itemID > 0 && createdAt > 0) {
        const [rows] = await db.query(
            "SELECT * FROM `items` WHERE `status` IN (?,?) AND category_id IN (?) AND (`created_at` < ? OR (`created_at` <= ? AND `id` < ?)) ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
            [
                ItemStatusOnSale,
                ItemStatusSoldOut,
                categoryIDs,
                new Date(createdAt),
                new Date(createdAt),
                itemID,
                ItemsPerPage + 1,
            ]
        );

        for (const row of rows) {
            items.push(row as Item);
        }
    } else {
        const [rows] = await db.query(
            "SELECT * FROM `items` WHERE `status` IN (?,?) AND category_id IN (?) ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
            [
                ItemStatusOnSale,
                ItemStatusSoldOut,
                categoryIDs,
                ItemsPerPage + 1,
            ]
        );

        for (const row of rows) {
            items.push(row as Item);
        }
    }

    let itemSimples: ItemSimple[] = [];

    for (const item of items) {
        const seller = await getUserSimpleByID(db, item.seller_id);
        if (seller === null) {
            replyError(reply, "seller not found", 404)
            await db.release();
            return;
        }
        const category = await getCategoryByID(db, item.category_id);
        if (category === null) {
            replyError(reply, "category not found", 404)
            await db.release();
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
        itemSimples = itemSimples.slice(0, itemSimples.length - 1)
    }

    const res = {
        root_category_id: rootCategory.id,
        root_category_name: rootCategory.category_name,
        items: itemSimples,
        has_next: hasNext,
    }

    await db.release();

    reply
        .code(200)
        .type("application/json")
        .send(res);

}

async function getTransactions(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const db = await getDBConnection();
    const user = await getLoginUser(req, db);

    if (user === null) {
        replyError(reply, "no session", 404);
        await db.release();
        return;
    }

    const query = req.query;
    let itemId = 0;
    if (query['item_id'] !== undefined) {
        itemId = parseInt(query['item_id'], 10);
        if (isNaN(itemId) || itemId <= 0) {
            replyError(reply, "item_id param error", 400);
            await db.release();
            return
        }
    }

    let createdAt = 0;
    if (query['created_at'] !== undefined) {
        createdAt = parseInt(query['created_at'], 10);
        if (isNaN(createdAt) || createdAt <= 0) {
            replyError(reply, "created_at param error", 400);
            await db.release();
            return
        }
    }

    await db.beginTransaction();
    const items: Item[] = [];
    if (itemId > 0 && createdAt > 0) {
        const [rows] = await db.query(
            "SELECT * FROM `items` WHERE (`seller_id` = ? OR `buyer_id` = ?) AND `status` IN (?,?,?,?,?) AND (`created_at` < ? OR (`created_at` <= ? AND `id` < ?)) ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
            [
                user.id,
                user.id,
                ItemStatusOnSale,
                ItemStatusTrading,
                ItemStatusSoldOut,
                ItemStatusCancel,
                ItemStatusStop,
                new Date(createdAt),
                new Date(createdAt),
                itemId,
                TransactionsPerPage + 1,
            ]
        );

        for (const row of rows) {
            items.push(row as Item);
        }

    } else {
        const [rows] = await db.query(
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
        const category = await getCategoryByID(db, item.category_id);
        if (category === null) {
            replyError(reply, "category not found", 404)
            await db.rollback();
            await db.release();
            return;
        }

        const seller = await getUserSimpleByID(db, item.seller_id);
        if (seller === null) {
            replyError(reply, "seller not found", 404)
            await db.rollback();
            await db.release();
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

        if (item.buyer_id !== undefined && item.buyer_id !== 0) {
            const buyer = await getUserSimpleByID(db, item.buyer_id);
            if (buyer === null) {
                replyError(reply, "buyer not found", 404);
                await db.rollback();
                await db.release();
                return;
            }
            itemDetail.buyer_id = item.buyer_id;
            itemDetail.buyer = buyer;
        }

        const [rows] = await db.query("SELECT * FROM `transaction_evidences` WHERE `item_id` = ?", [item.id]);
        let transactionEvidence: TransactionEvidence | null = null;
        for (const row of rows) {
            transactionEvidence = row as TransactionEvidence;
        }

        if (transactionEvidence !== null) {
            const [rows] = await db.query("SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ?", [transactionEvidence.id]);

            let shipping: Shipping | null = null;
            for (const row of rows) {
                shipping = row as Shipping;
            }

            if (shipping === null) {
                replyError(reply, "shipping not found", 404);
                await db.rollback();
                await db.release();
                return;
            }

            try {
                const res = await shipmentStatus(await getShipmentServiceURL(db), {reserve_id: shipping.reserve_id});
                itemDetail.shipping_status = res.status;
            } catch (error) {
                replyError(reply, "failed to request to shipment service");
                await db.rollback();
                await db.release();
                return;
            }

            itemDetail.transaction_evidence_id = transactionEvidence.id;
            itemDetail.transaction_evidence_status = transactionEvidence.status;
        }

        itemDetails.push(itemDetail);

    }

    await db.commit();

    let hasNext = false;
    if (itemDetails.length > TransactionsPerPage) {
        hasNext = true;
        itemDetails = itemDetails.slice(0, TransactionsPerPage);
    }

    await db.release();

    reply
        .code(200)
        .type("application/json;charset=utf-8")
        .send({has_next: hasNext, items: itemDetails});

}

async function getUserItems(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const userIdStr = req.params.user_id;
    const userId = parseInt(userIdStr, 10);
    if (userId === undefined || isNaN(userId)) {
        replyError(reply, "incorrect user id", 400);
        return;
    }

    const db = await getDBConnection();
    const userSimple = await getUserSimpleByID(db, userId);
    if (userSimple === null) {
        replyError(reply, "user not found", 404);
        await db.release();
        return;
    }

    const itemIDStr = req.query.item_id;
    let itemID = 0;
    if (itemIDStr !== undefined && itemIDStr !== "") {
        itemID = parseInt(itemIDStr, 10);
        if (isNaN(itemID) || itemID <= 0) {
            replyError(reply, "item_id param error", 400);
            await db.release();
            return;
        }
    }
    const createdAtStr = req.query.created_at;
    let createdAt = 0;
    if (createdAtStr !== undefined && createdAtStr !== "") {
        createdAt = parseInt(createdAtStr, 10);
        if (isNaN(createdAt) || createdAt <= 0) {
            replyError(reply, "created_at param error", 400);
            await db.release();
            return;
        }
    }

    const items: Item[] = [];
    if (itemID > 0 && createdAt > 0) {
        const [rows] = await db.query(
            "SELECT * FROM `items` WHERE `seller_id` = ? AND `status` IN (?,?,?) AND (`created_at` < ? OR (`created_at` <= ? AND `id` < ?)) ORDER BY `created_at` DESC, `id` DESC LIMIT ?",
            [
                userSimple.id,
                ItemStatusOnSale,
                ItemStatusTrading,
                ItemStatusSoldOut,
                new Date(createdAt),
                new Date(createdAt),
                itemID,
                ItemsPerPage + 1,
            ]
        );

        for (const row of rows) {
            items.push(row as Item);
        }
    } else {
        const [rows] = await db.query(
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
        const category = await getCategoryByID(db, item.category_id);
        if (category === null) {
            replyError(reply, "category not found", 404)
            await db.release();
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

    await db.release();

    reply
        .code(200)
        .type("application/json")
        .send(res);
}

async function getItem(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const itemIdStr = req.params.item_id;
    const itemId = parseInt(itemIdStr, 10);
    if (itemId === undefined || isNaN(itemId)) {
        replyError(reply, "incorrect item id", 400);
        return;
    }

    const db = await getDBConnection();
    const user = await getLoginUser(req, db);
    if (user === null) {
        replyError(reply, "no session", 404);
        await db.release();
        return;
    }

    const [rows] = await db.query("SELECT * FROM `items` WHERE `id` = ?", [itemId]);
    let item: Item | null = null;

    for (const row of rows) {
        item = row as Item;
    }

    if (item === null) {
        replyError(reply, "item not found", 404);
        await db.release();
        return;
    }

    const category = await getCategoryByID(db, item.category_id);
    if (category === null) {
        replyError(reply, "category not found", 404)
        await db.release();
        return;
    }

    const seller = await getUserSimpleByID(db, item.seller_id);
    if (seller === null) {
        replyError(reply, "seller not found", 404)
        await db.release();
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

    if ((user.id === item.seller_id || user.id === item.buyer_id) && item.buyer_id !== 0) {
        const buyer = await getUserSimpleByID(db, item.buyer_id);
        if (buyer === null) {
            replyError(reply, "buyer not found", 404);
            await db.release();
            return;
        }

        itemDetail.buyer_id = item.buyer_id;
        itemDetail.buyer = buyer;

        const [rows] = await db.query("SELECT * FROM `transaction_evidences` WHERE `item_id` = ?", [item.id]);
        let transactionEvidence: TransactionEvidence | null = null;
        for (const row of rows) {
            transactionEvidence = row as TransactionEvidence;
        }

        if (transactionEvidence !== null) {
            const [rows] = await db.query("SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ?", [transactionEvidence.id])
            let shipping: Shipping | null = null;
            for (const row of rows) {
                shipping = row as Shipping;
            }

            if (shipping === null) {
                replyError(reply, "shipping not found", 404);
                await db.release();
                return;
            }

            itemDetail.transaction_evidence_id = transactionEvidence.id;
            itemDetail.transaction_evidence_status = transactionEvidence.status;
            itemDetail.shipping_status = shipping.status;
        }

    }

    await db.release();

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
        replyError(reply, "csrf token error", 422)
        return;
    }

    if (price < ItemMinPrice || price > ItemMaxPrice) {
        replyError(reply, ItemPriceErrMsg, 400);
        return;
    }

    const db = await getDBConnection();

    const seller = await getLoginUser(req, db);
    if (seller === null) {
        replyError(reply, "no session", 404);
        await db.release();
        return;
    }

    let targetItem: Item | null = null;
    ;
    {
        const [rows] = await db.query("SELECT * FROM `items` WHERE `id` = ?", [itemID]);
        for (const row of rows) {
            targetItem = row as Item;
        }
    }

    if (targetItem === null) {
        replyError(reply, "item not found");
        await db.release();
        return;
    }

    if (targetItem.seller_id !== seller.id) {
        replyError(reply, "自分の商品以外は編集できません", 403);
        await db.release();
        return;
    }

    await db.beginTransaction();

    await db.query("SELECT * FROM `items` WHERE `id` = ? FOR UPDATE", [targetItem.id]);

    if (targetItem.status !== ItemStatusOnSale) {
        replyError(reply, "販売中の商品以外編集できません", 403);
        await db.rollback();
        return;
    }

    await db.query("UPDATE `items` SET `price` = ?, `updated_at` = ? WHERE `id` = ?", [price, new Date(), targetItem.id]);

    {
        const [rows] = await db.query("SELECT * FROM `items` WHERE `id` = ?", [targetItem.id]);
        for (const row of rows) {
            targetItem = row as Item;
        }
    }

    await db.commit();
    await db.release();

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
        replyError(reply, "csrf token error", 422);
        return;
    }

    const db = await getDBConnection();

    const buyer = await getLoginUser(req, db);

    if (buyer === null) {
        replyError(reply, "no session", 404);
        await db.release();
        return;
    }

    await db.beginTransaction();

    let targetItem: Item | null = null;
    {
        const [rows] = await db.query("SELECT * FROM `items` WHERE `id` = ? FOR UPDATE", [req.body.item_id]);

        for (const row of rows) {
            targetItem = row as Item;
        }
    }

    if (targetItem === null) {
        replyError(reply, "item not found", 404);
        await db.rollback();
        await db.release();
        return;
    }

    if (targetItem.status !== ItemStatusOnSale) {
        replyError(reply, "item is not for sale", 403);
        await db.rollback();
        await db.release();
        return;
    }

    if (targetItem.seller_id === buyer.id) {
        replyError(reply, "自分の商品は買えません", 403);
        await db.rollback();
        await db.release();
        return;
    }

    let seller: User | null = null;
    {
        const [rows] = await db.query("SELECT * FROM `users` WHERE `id` = ? FOR UPDATE", [targetItem.seller_id]);
        for (const row of rows) {
            seller = row as User;
        }
    }

    if (seller === null) {
        replyError(reply, "seller not found", 404);
        await db.rollback();
        await db.release();
        return;
    }

    const category = await getCategoryByID(db, targetItem.category_id);
    if (category === null) {
        replyError(reply, "category id error", 500);
        await db.rollback();
        await db.release();
        return;
    }

    const [result] = await db.query(
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

    await db.query(
        "UPDATE `items` SET `buyer_id` = ?, `status` = ?, `updated_at` = ? WHERE `id` = ?",
        [
            buyer.id,
            ItemStatusTrading,
            new Date(),
            targetItem.id,
        ]
    )

    try {
        const scr = await shipmentCreate(await getShipmentServiceURL(db), {
            to_address: buyer.address,
            to_name: buyer.account_name,
            from_address: seller.address,
            from_name: seller.account_name,
        });

        try {
            const pstr = await paymentToken(await getPaymentServiceURL(db), {
                shop_id: PaymentServiceIsucariShopID.toString(),
                token: req.body.token,
                api_key: PaymentServiceIsucariAPIKey,
                price: targetItem.price,
            });

            if (pstr.status === "invalid") {
                replyError(reply, "カード情報に誤りがあります", 400);
                await db.rollback();
                await db.release();
                return;
            }
            if (pstr.status === "fail") {
                replyError(reply, "カードの残高が足りません", 400);
                await db.rollback();
                await db.release();
                return;
            }

            if (pstr.status !== 'ok') {
                replyError(reply, "想定外のエラー", 400)
                await db.rollback()
                await db.release();
                return;
            }

            await db.query(
                "INSERT INTO `shippings` (`transaction_evidence_id`, `status`, `item_name`, `item_id`, `reserve_id`, `reserve_time`, `to_address`, `to_name`, `from_address`, `from_name`, `img_binary`) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
                [
                    transactionEvidenceId,
                    ShippingsStatusInitial,
                    targetItem.name,
                    targetItem.id,
                    scr.reserve_id,
                    scr.reserve_time,
                    buyer.address,
                    buyer.account_name,
                    seller.address,
                    seller.account_name,
                    "",
                ]
            );
        } catch (e) {
            replyError(reply, "payment service is failed", 500)
            await db.rollback();
            await db.release();
            return;
        }
    } catch (error) {
        replyError(reply, "failed to request to shipment service", 500);
        await db.rollback();
        await db.release();
        return;
    }

    await db.commit();
    await db.release();

    reply.code(200)
        .type("application/json;charset=utf-8")
        .send({
            transaction_evidence_id: transactionEvidenceId,
        });

}

async function postSell(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const csrfToken = req.body.csrf_token;
    const name = req.body.name;
    const description = req.body.description;
    const priceStr = req.body.price;
    const categoryIdStr = req.body.category_id;

    if (csrfToken !== req.cookies.csrf_token) {
        replyError(reply, "csrf token error", 422);
        return;
    }

    const categoryId: number = parseInt(categoryIdStr, 10);
    if (isNaN(categoryId) || categoryId < 0) {
        replyError(reply, "category id error", 400);
        return;
    }

    const price: number = parseInt(priceStr, 10);
    if (isNaN(price) || price < 0) {
        replyError(reply, "price error", 400);
        return;
    }

    if (price < ItemMinPrice || price > ItemMaxPrice) {
        replyError(reply, ItemPriceErrMsg, 400);
        return;
    }

    if (name === null || name === "" || description === null || description === "" || price === 0 || categoryId === 0) {
        replyError(reply, "all parameters are required", 400);
    }

    const db = await getDBConnection();

    const category = await getCategoryByID(db, categoryId);
    if (category === null || category.parent_id === 0) {
        replyError(reply, "Incorrect category ID", 400);
        await db.release();
        return;
    }

    const user = await getLoginUser(req, db);

    if (user === null) {
        replyError(reply, "no session", 404);
        await db.release();
        return;
    }

    let ext = path.extname(req.body.image[0].filename);
    if (![".jpg", ".jpeg", ".png", ".gif"].includes(ext)) {
        replyError(reply, "unsupported image format error", 400);
        await db.release();
        return;
    }

    if (ext === ".jpeg") {
        ext = ".jpg";
    }


    const imgName = `${await getRandomString(16)}${ext}`;

    await fs.promises.writeFile(`../public/upload/${imgName}`, req.body.image[0].data);

    await db.beginTransaction();

    let seller: User | null = null;
    {
        const [rows] = await db.query("SELECT * FROM `users` WHERE `id` = ? FOR UPDATE", [user.id]);
        for (const row of rows) {
            seller = row as User;
        }
    }

    if (seller === null) {
        replyError(reply, "user not found", 404);
        await db.rollback();
        await db.release();
        return;
    }

    const [result] = await db.query("INSERT INTO `items` (`seller_id`, `status`, `name`, `price`, `description`,`image_name`,`category_id`) VALUES (?, ?, ?, ?, ?, ?, ?)", [
        seller.id,
        ItemStatusOnSale,
        name,
        price,
        description,
        imgName,
        category.id,
    ]);

    const itemId = result.insertId;

    const now = new Date();
    await db.query("UPDATE `users` SET `num_sell_items`=?, `last_bump`=? WHERE `id`=?", [
        seller.num_sell_items + 1,
        now,
        seller.id,
    ]);

    await db.commit();
    await db.release();

    reply
        .code(200)
        .type("application/json;charset=utf-8")
        .send({
            id: itemId,
        });

}

async function postShip(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const csrfToken = req.body.csrf_token;
    const itemId = req.body.item_id;

    if (csrfToken !== req.cookies.csrf_token) {
        replyError(reply, "csrf token error", 422);
        return;
    }

    const db = await getDBConnection();

    const seller = await getLoginUser(req, db);

    if (seller === null) {
        replyError(reply, "no session", 404);
        await db.release();
        return;
    }

    let transactionalEvidence: TransactionEvidence | null = null;
    {
        const [rows] = await db.query(
            "SELECT * FROM `transaction_evidences` WHERE `item_id` = ?",
            [itemId]
        )

        for (const row of rows) {
            transactionalEvidence = row as TransactionEvidence;
        }

    }

    if (transactionalEvidence === null) {
        replyError(reply, "transaction_evidences not found", 404);
        await db.release();
        return;
    }

    if (transactionalEvidence.seller_id !== seller.id) {
        replyError(reply, "権限がありません", 403);
        await db.release();
        return;
    }

    await db.beginTransaction();

    let item: Item | null = null;
    {
        const [rows] = await db.query(
            "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE",
            [itemId]
        );
        for (const row of rows) {
            item = row as Item;
        }
    }

    if (item === null) {
        replyError(reply, "item not found", 404);
        await db.rollback();
        await db.release();
        return;
    }

    if (item.status !== ItemStatusTrading) {
        replyError(reply, "アイテムが取引中ではありません", 403);
        await db.rollback();
        await db.release();
        return;
    }

    {
        const [rows] = await db.query(
            "SELECT * FROM `transaction_evidences` WHERE `id` = ? FOR UPDATE",
            [
                transactionalEvidence.id,
            ]
        )
        if (rows.length === 0) {
            replyError(reply, "transaction_evidences not found", 404);
            await db.rollback();
            await db.release();
            return;
        }
    }

    if (transactionalEvidence.status !== TransactionEvidenceStatusWaitShipping) {
        replyError(reply, "準備ができていません", 403);
        await db.rollback();
        await db.release();
        return;
    }

    let shipping: Shipping | null = null;
    {
        const [rows] = await db.query(
            "SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE",
            [
                transactionalEvidence.id,
            ]
        );

        for (const row of rows) {
            shipping = row as Shipping;
        }
    }

    if (shipping === null) {
        replyError(reply, "shippings not found", 404);
        await db.rollback();
        await db.release();
        return;
    }

    const img = await shipmentRequest(await getShipmentServiceURL(db), {
        reserve_id: shipping.reserve_id,
    });

    await db.query(
        "UPDATE `shippings` SET `status` = ?, `img_binary` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?",
        [
            ShippingsStatusWaitPickup,
            img,
            new Date(),
            transactionalEvidence.id,
        ]
    );

    await db.commit();
    await db.release();

    reply
        .code(200)
        .type("application/json")
        .send({
            path: `/transactions/${transactionalEvidence.id}.png`,
            reserve_id: shipping.reserve_id,
        });

}

async function postShipDone(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const csrfToken = req.body.csrf_token;
    const itemId = req.body.item_id;

    if (csrfToken !== req.cookies.csrf_token) {
        replyError(reply, "csrf token error", 422);
        return;
    }

    const db = await getDBConnection();

    const seller = await getLoginUser(req, db)

    if (seller === null) {
        replyError(reply, "no session", 404);
        await db.release();
        return;
    }

    let transactionEvidence: TransactionEvidence | null = null;
    {
        const [rows] = await db.query(
            "SELECT * FROM `transaction_evidences` WHERE `item_id` = ?",
            [
                itemId,
            ]
        );
        for (const row of rows) {
            transactionEvidence = row as TransactionEvidence;
        }
    }

    if (transactionEvidence === null) {
        replyError(reply, "transaction_evidence not found", 404);
        await db.release();
        return;
    }

    if (transactionEvidence.seller_id !== seller.id) {
        replyError(reply, "権限がありません", 403);
        await db.release();
        return;
    }

    await db.beginTransaction();

    let item: Item | null = null;
    {
        const [rows] = await db.query("SELECT * FROM `items` WHERE `id` = ? FOR UPDATE", [
            itemId,
        ]);

        for (const row of rows) {
            item = row as Item;
        }

    }

    if (item === null) {
        replyError(reply, "items not found", 404);
        await db.rollback();
        await db.release();
        return;
    }

    if (item.status !== ItemStatusTrading) {
        replyError(reply, "商品が取引中ではありません", 403);
        await db.rollback();
        await db.release();
        return;
    }

    {
        const [rows] = await db.query(
            "SELECT * FROM `transaction_evidences` WHERE `id` = ? FOR UPDATE",
            [
                transactionEvidence.id,
            ]
        )
        for (const row of rows) {
            transactionEvidence = row as TransactionEvidence;
        }
    }

    if (transactionEvidence === null) {
        replyError(reply, "transaction_evidences not found", 404);
        await db.rollback();
        await db.release();
        return;
    }

    if (transactionEvidence.status !== TransactionEvidenceStatusWaitShipping) {
        replyError(reply, "準備ができていません", 403);
        await db.rollback();
        await db.release();
        return;
    }

    let shipping: Shipping | null = null;
    {
        const [rows] = await db.query(
            "SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE",
            [
                transactionEvidence.id,
            ]
        )

        for (const row of rows) {
            shipping = row as Shipping;
        }
    }

    if (shipping === null) {
        replyError(reply, "shippings not found", 404);
        await db.rollback();
        await db.release();
        return;
    }

    let params = {
        reserve_id: shipping.reserve_id,
    }
    try {
        const res = await shipmentStatus(await getShipmentServiceURL(db), params)
        if (!(res.status === ShippingsStatusShipping || res.status === ShippingsStatusDone)) {
            replyError(reply, "shipment service側で配送中か配送完了になっていません", 403);
            await db.rollback();
            await db.release();
            return;
        }

        await db.query(
            "UPDATE `shippings` SET `status` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?",
            [
                res.status,
                new Date(),
                transactionEvidence.id,
            ]
        );

    } catch (res) {
        replyError(reply, "failed to request to shipment service");
        await db.rollback();
        await db.release();
        return;
    }

    await db.query(
        "UPDATE `transaction_evidences` SET `status` = ?, `updated_at` = ? WHERE `id` = ?",
        [
            TransactionEvidenceStatusWaitDone,
            new Date(),
            transactionEvidence.id,
        ]
    );

    await db.commit();
    await db.release();

    reply
        .code(200)
        .type("application/json;charset=utf-8")
        .send({
            transaction_evidence_id: transactionEvidence.id,
        });

}

async function postComplete(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const csrfToken = req.body.csrf_token;
    const itemId = req.body.item_id;

    if (csrfToken !== req.cookies.csrf_token) {
        replyError(reply, "csrf token error", 422);
        return;
    }

    const db = await getDBConnection();
    const buyer = await getLoginUser(req, db);

    if (buyer === null) {
        replyError(reply, "no session", 404);
        await db.release();
        return;
    }

    let transactionEvidence: TransactionEvidence | null = null;
    {
        const [rows] = await db.query("SELECT * FROM `transaction_evidences` WHERE `item_id` = ?", [itemId])
        for (const row of rows) {
            transactionEvidence = row as TransactionEvidence;
        }
    }

    if (transactionEvidence === null) {
        replyError(reply, "transaction_evidence not found", 404);
        await db.release();
        return;
    }

    if (transactionEvidence.buyer_id !== buyer.id) {
        replyError(reply, "権限がありません", 403);
        await db.release();
        return;
    }

    await db.beginTransaction();

    let item: Item | null = null;
    {
        const [rows] = await db.query("SELECT * FROM `items` WHERE `id` = ? FOR UPDATE", [itemId])
        for (const row of rows) {
            item = row as Item;
        }
    }

    if (item === null) {
        replyError(reply, "items not found", 404);
        await db.rollback();
        await db.release();
        return;
    }

    if (item.status !== ItemStatusTrading) {
        replyError(reply, "商品が取引中ではありません", 403);
        await db.rollback();
        await db.release();
        return;
    }

    {
        const [rows] = await db.query("SELECT * FROM `transaction_evidences` WHERE `item_id` = ? FOR UPDATE", [itemId])
        for (const row of rows) {
            transactionEvidence = row as TransactionEvidence;
        }
    }

    if (transactionEvidence === null) {
        replyError(reply, "transaction_evidences not found", 404);
        await db.rollback();
        await db.release();
        return;
    }

    if (transactionEvidence.status !== TransactionEvidenceStatusWaitDone) {
        replyError(reply, "準備ができていません", 403);
        await db.rollback();
        await db.release();
        return;
    }

    let shipping: Shipping | null = null;
    {
        const [rows] = await db.query("SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE", [transactionEvidence.id]);
        for (const row of rows) {
            shipping = row as Shipping;
        }
    }

    if (shipping === null) {
        replyError(reply, "shipping not found", 404);
        await db.rollback();
        await db.release();
        return;
    }

    try {
        const res = await shipmentStatus(await getShipmentServiceURL(db), {
            reserve_id: shipping.reserve_id,
        })
        if (res.status !== ShippingsStatusDone) {
            replyError(reply, "shipment service側で配送完了になっていません", 400);
            await db.rollback();
            await db.release();
            return;
        }
    } catch (e) {
        replyError(reply, "failed to request to shipment service", 500);
        await db.rollback();
        await db.release();
        return;

    }

    await db.query("UPDATE `shippings` SET `status` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?", [
        ShippingsStatusDone,
        new Date(),
        transactionEvidence.id,
    ])

    await db.query("UPDATE `transaction_evidences` SET `status` = ?, `updated_at` = ? WHERE `id` = ?", [
        TransactionEvidenceStatusDone,
        new Date(),
        transactionEvidence.id,
    ]);

    await db.query("UPDATE `items` SET `status` = ?, `updated_at` = ? WHERE `id` = ?", [
        ItemStatusSoldOut,
        new Date(),
        itemId,
    ]);

    await db.commit();
    await db.release();

    reply
        .code(200)
        .type("application/json;charset=utf-8")
        .send({
            transaction_evidence_id: transactionEvidence.id,
        });

}

async function getQRCode(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const transactionEvidenceIdStr: string = req.params.transaction_evidence_id;
    const transactionEvidenceId: number = parseInt(transactionEvidenceIdStr, 10);
    if (transactionEvidenceId === null || isNaN(transactionEvidenceId)) {
        replyError(reply, "incorrect transaction_evidence id", 400);
        return;
    }

    const db = await getDBConnection();
    const seller = await getLoginUser(req, db);
    if (seller === null) {
        replyError(reply, "no session", 404);
        await db.release();
        return;
    }

    let transactionEvidence: TransactionEvidence | null = null;
    {
        const [rows] = await db.query("SELECT * FROM `transaction_evidences` WHERE `id` = ?", [transactionEvidenceId]);
        for (const row of rows) {
            transactionEvidence = row as TransactionEvidence;
        }
    }

    if (transactionEvidence === null) {
        replyError(reply, "transaction_evidence not found", 404);
        await db.release();
        return;
    }

    if (transactionEvidence.seller_id !== seller.id) {
        replyError(reply, "権限がありません", 403);
        await db.release();
        return;
    }

    let shipping: Shipping | null = null;
    {
        const [rows] = await db.query("SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ?", [transactionEvidence.id]);
        for (const row of rows) {
            shipping = row as Shipping;
        }
    }

    if (shipping === null) {
        replyError(reply, "shippings not found", 404);
        await db.release();
        return;
    }

    if (shipping.status !== ShippingsStatusWaitPickup && shipping.status !== ShippingsStatusShipping) {
        replyError(reply, "qrcode not available", 403);
        await db.release();
        return;
    }

    if (shipping.img_binary.byteLength === 0) {
        replyError(reply, "empty qrcode image")
        await db.release();
        return;
    }

    await db.release();

    reply
        .code(200)
        .type("image/png")
        .send(shipping.img_binary);

}

async function postBump(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const csrfToken = req.body.csrf_token;
    const itemId = req.body.item_id;

    if (csrfToken !== req.cookies.csrf_token) {
        replyError(reply, "csrf token error", 422);
        return;
    }

    const db = await getDBConnection();

    const user = await getLoginUser(req, db);
    if (user === null) {
        replyError(reply, "no session", 404);
        await db.release();
        return;
    }


    await db.beginTransaction();

    let targetItem: Item | null = null;
    {
        const [rows] = await db.query(
            "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE",
            [
                itemId,
            ]
        )
        for (const row of rows) {
            targetItem = row as Item;
        }
    }

    if (targetItem === null) {
        replyError(reply, "item not found", 404);
        await db.rollback();
        await db.release();
        return;
    }

    if (targetItem.seller_id !== user.id) {
        replyError(reply, "自分の商品以外は編集できません", 403);
        await db.rollback();
        await db.release();
        return;
    }

    let seller: User | null = null;
    {
        const [rows] = await db.query(
            "SELECT * FROM `users` WHERE `id` = ? FOR UPDATE",
            [
                user.id,
            ]
        );
        for (const row of rows) {
            seller = row as User;
        }
    }

    if (seller === null) {
        replyError(reply, "user not found", 404);
        await db.rollback();
        await db.release();
        return;
    }

    // last bump + 3s > 0
    const now = new Date();
    if (seller.last_bump.getTime() + BumpChargeSeconds > now.getTime()) {
        replyError(reply, "Bump not allowed", 403)
        await db.rollback();
        await db.release();
        return;
    }

    await db.query(
        "UPDATE `items` SET `created_at`=?, `updated_at`=? WHERE id=?",
        [
            now,
            now,
            targetItem.id,
        ]
    );

    await db.query("UPDATE `users` SET `last_bump`=? WHERE id=?", [now, seller.id])

    {
        const [rows] = await db.query("SELECT * FROM `items` WHERE `id` = ?", [itemId]);
        for (const row of rows) {
            targetItem = row as Item;
        }
    }

    await db.commit();
    await db.release();

    reply
        .code(200)
        .type("application/json;charset=utf-8")
        .send({
            item_id: targetItem.id,
            item_price: targetItem.price,
            item_created_at: targetItem.created_at.getTime(),
            item_updated_at: targetItem.updated_at.getTime(),
        });

}

async function getSettings(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
    const csrfToken = req.cookies.csrf_token;

    const db = await getDBConnection();
    const user = await getLoginUser(req, db);

    const res = {
        user: null as User | null,
        payment_service_url: null as string | null,
        categories: null as Category[] | null,
        csrf_token: null as string | null,
    };

    res.user = user;
    res.payment_service_url = await getPaymentServiceURL(db);
    res.csrf_token = csrfToken;

    const categories: Category[] = [];
    const [rows] = await db.query("SELECT * FROM `categories`", []);
    for (const row of rows) {
        categories.push(row as Category);
    }
    res.categories = categories;

    await db.release();

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

        replyError(reply, "all parameters are required", 400);
        return;
    }

    const db = await getDBConnection();
    const [rows] = await db.query("SELECT * FROM `users` WHERE `account_name` = ?", [accountName])
    let user: User | null = null;
    for (const row of rows) {
        user = row as User;
    }

    if (user === null) {
        replyError(reply, "アカウント名かパスワードが間違えています", 401);
        await db.release();
        return;
    }

    if (!await comparePassword(password, user.hashed_password)) {
        replyError(reply, "アカウント名かパスワードが間違えています", 401);
        await db.release();
        return;
    }

    reply.setCookie("user_id", user.id.toString(), {
        path: "/",
    });
    reply.setCookie("csrf_token", await getRandomString(128), {
        path: "/",
    });

    await db.release();

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
        replyError(reply, "all parameters are required", 400);
        return;
    }

    const db = await getDBConnection();

    const [rows] = await db.query(
        "SELECT * FROM `users` WHERE `account_name` = ?",
        [
            accountName,
        ]
    );

    if (rows.length > 0) {
        replyError(reply, "アカウント名かパスワードが間違えています", 401);
        await db.release();
        return;
    }

    const hashedPassword = await encryptPassword(password);

    const [result,] = await db.query(
        "INSERT INTO `users` (`account_name`, `hashed_password`, `address`) VALUES (?, ?, ?)",
        [
            accountName,
            hashedPassword,
            address,
        ]
    );

    await db.release();

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
    const db = await getDBConnection();
    const [rows] = await db.query("SELECT * FROM `transaction_evidences` WHERE `id` > 15007");
    const transactionEvidences: TransactionEvidence[] = [];
    for (const row of rows) {
        transactionEvidences.push(row as TransactionEvidence);
    }

    await db.release();

    reply
        .code(200)
        .type("application/json")
        .send(transactionEvidences);
}

async function getLoginUser(req: FastifyRequest, db: MySQLQueryable): Promise<User | null> {
    let userId: number;
    if (req.cookies.user_id !== undefined && req.cookies.user_id !== "") {
        const [rows] = await db.query("SELECT * FROM `users` WHERE `id` = ?", [req.cookies.user_id]);
        for (const row of rows) {
            const user = row as User;
            return user;
        }
    }

    return null;
}

function getSession(req: FastifyRequest) {
}

fastify.listen(8000, (err, _address) => {
    if (err) {
        throw new TraceError("Failed to listening", err);
    }
});

function replyError(reply: FastifyReply<ServerResponse>, message: string, status = 500) {
    reply.code(status)
        .type("application/json")
        .send({"error": message});
}

async function getUserSimpleByID(db: MySQLQueryable, userID: number): Promise<UserSimple | null> {
    const [rows,] = await db.query("SELECT * FROM `users` WHERE `id` = ?", [userID]);
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

async function getCategoryByID(db: MySQLQueryable, categoryId: number): Promise<Category | null> {
    const [rows,] = await db.query("SELECT * FROM `categories` WHERE `id` = ?", [categoryId]);
    for (const row of rows) {
        const category = row as Category;
        if (category.parent_id !== undefined && category.parent_id != 0) {
            const parentCategory = await getCategoryByID(db, category.parent_id);
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

async function getConfigByName(db: MySQLQueryable, name: string): Promise<string | null> {
    let config: Config | null = null;
    {
        const [rows] = await db.query("SELECT * FROM `configs` WHERE `name` = ?", [name]);
        for (const row of rows) {
            config = row as Config;
        }
    }

    if (config === null) {
        return null;
    }

    return config.val;
}

async function getPaymentServiceURL(db: MySQLQueryable): Promise<string> {
    const result = await getConfigByName(db, "payment_service_url");
    if (result === null) {
        return DefaultPaymentServiceURL;
    }
    return result;
}

async function getShipmentServiceURL(db: MySQLQueryable): Promise<string> {
    const result = await getConfigByName(db, "shipment_service_url");
    if (result === null) {
        return DefaultShipmentServiceURL;
    }
    return result;
}

function getImageURL(imageName: string) {
    return `/upload/${imageName}`;
}
