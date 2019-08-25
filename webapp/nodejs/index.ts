import { IncomingMessage, ServerResponse } from "http";
import util from "util";
import childProcess from "child_process";
import path from "path";
import fs from "fs";

import TraceError from "trace-error";
import createFastify, { FastifyRequest, FastifyReply } from "fastify";
// @ts-ignore
import fastifyMysql from "fastify-mysql";
import fastifyCookie from "fastify-cookie";
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

type ReqInitialize = Readonly<{
  payment_service_url: string;
  shipment_service_url: string;
}>;

const fastify = createFastify({
  logger: true
});

fastify.register(fastifyStatic, {
  root: path.join(__dirname, "public")
});

fastify.register(fastifyMysql, {
  host: process.env.DB_HOST || "127.0.0.1",
  port: process.env.DB_PORT || "3306",
  user: process.env.DB_USER || "isucari",
  password: process.env.DB_PASS || "isucari",
  database: process.env.DB_DATABASE || "isucari",

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

fastify.post("/initialize", async (req, reply) => {
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
});

fastify.get("/new_items.json", (_req, reply) => {
  TODO();
});

fastify.get("/new_items/:root_category_id.json", (req, reply) => {
  const rootCategoryId: string = req.params.root_category_id;
  TODO();
});

fastify.get("/users/transactions.json", (req, reply) => {
  TODO();
});

fastify.get("/users/:user_id.json", (req, reply) => {
  const userId: string = req.params.user_id;
  TODO();
});

fastify.get("/items/:item_id.json", (req, reply) => {
  const itemId: string = req.params.item_id;
  TODO();
});

fastify.post("/items/edit", (req, reply) => {
  TODO();
});

fastify.post("/buy", (req, reply) => {
  TODO();
});

fastify.post("/sell", (req, reply) => {
  TODO();
});

fastify.post("/ship", (req, reply) => {
  TODO();
});

fastify.post("/ship_done", (req, reply) => {
  TODO();
});

fastify.post("/complete", (req, reply) => {
  TODO();
});

fastify.get("/transactions/:transaction_evidence_id.png", (req, reply) => {
  const transactionEvidenceId: string = req.params.transaction_evidence_id;
  TODO();
});

fastify.post("/bump", (req, reply) => {
  TODO();
});

fastify.get("/settings", (req, reply) => {
  TODO();
});

fastify.post("/login", (req, reply) => {
  TODO();
});

fastify.post("/register", (req, reply) => {
  TODO();
});

fastify.get("/reports.json", (req, reply) => {
  TODO();
});

// Frontend

async function getIndex(_req: any, reply: FastifyReply<ServerResponse>) {
  const html = await fs.promises.readFile(path.join(__dirname, "public/index.html"));
  reply.type("text/html").send(html);
}

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

fastify.listen(8000, (err, _address) => {
  if (err) {
    throw new TraceError("Failed to listening", err);
  }
});
