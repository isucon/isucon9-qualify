import { IncomingMessage } from "http";
import util from "util";
import childProcess from "child_process";

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

type ReqInitialize = Readonly<{
  payment_service_url: string;
  shipment_service_url: string;
}>;

const fastify = createFastify({
  logger: true
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

fastify.get("/hello", (_req, reply) => {
  reply.send("Hello, world!\n");
});

fastify.listen(8000, (err, _address) => {
  if (err) {
    throw new TraceError("Failed to listening", err);
  }
});
