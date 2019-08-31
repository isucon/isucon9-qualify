import { IncomingMessage, ServerResponse } from "http";
import util, {types} from "util";
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
  creaed_at: Date;
  updated_at: Date;
};

type ItemSimple = {
  id: number;
  seller_id: number;
  seller: UserSimple;
  status: string;
  name: string;
  price: number;
  description: string;
  image_url: string;
  category_id: number;
  category: Category;
  creaed_at: Date;
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

type Category = {};

type ReqInitialize = {
  payment_service_url: string;
  shipment_service_url: string;
};

const fastify = createFastify({
  logger: true
});

fastify.register(fastifyStatic, {
  root: path.join(__dirname, "public")
});

fastify.register(fastifyCookie);
fastify.register(fastifySession, {secret: '123456789012345678901234567890123'});

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

fastify.post("/initialize", getInitialize);
fastify.get("/new_items.json", getNewItems);
fastify.get("/new_items/:root_category_id.json", getNewCategoryItems);
fastify.get("/users/transactions.json", getTransactions);
fastify.get("/users/:user_id.json", getUserItems);
fastify.get("/items/:item_id.json", getItem);
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

async function getInitialize(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
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
}

async function getNewCategoryItems(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
  const rootCategoryId: string = req.params.root_category_id;
}

async function getTransactions(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
}

async function getUserItems(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
}

async function getItem(req: FastifyRequest, reply: FastifyReply<ServerResponse>) {
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

  const res = { User: null as User | null, PaymentServiceURL: null as string | null, Categories: null as Category[] | null};
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
}

function getCsrfToken(req: FastifyRequest) {

}

async function getUser(req: FastifyRequest): Promise<User|null> {
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

