import { serve } from '@hono/node-server';
import { serveStatic } from '@hono/node-server/serve-static';
import { Hono, Context } from 'hono';
import { HTTPException } from 'hono/http-exception';
import type { ContentfulStatusCode } from 'hono/utils/http-status';
import { readFile, writeFile } from 'fs/promises';
import { join, extname } from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import { exec } from 'child_process';
import { promisify } from 'util';

import {
  User,
  Item,
  ItemSimple,
  ItemDetail,
  Category,
  TransactionEvidence,
  Shipping,
  ReqInitialize,
  ReqRegister,
  ReqLogin,
  ResNewItems,
  ResUserItems,
  ResTransactions,
  ReqItemEdit,
  ReqBuy,
  ReqSell,
  ReqShip,
  ReqShipDone,
  ReqComplete,
  ReqBump,
  ResSettings,
  ResTransactionEvidence,
  ResItemEdit,
  ResSell,
  ResShip,
  ResBump,
  ResLogin,
  ResRegister,
  ResInitialize,
} from './types.js';

import {
  ITEM_MIN_PRICE,
  ITEM_MAX_PRICE,
  ITEM_PRICE_ERR_MSG,
  ITEM_STATUS,
  SHIPPINGS_STATUS,
  BUMP_CHARGE_SECONDS,
  ITEMS_PER_PAGE,
  TRANSACTIONS_PER_PAGE,
  SHIPMENT_SERVICE_ISUCARI_API_KEY,
} from './constants.js';

import {
  secureRandomStr,
  hashPassword,
  verifyPassword,
  getImageURL,
} from './utils.js';

import {
  getUserSimpleByID,
  getCategoryByID,
  getPaymentServiceURL,
  getShipmentServiceURL,
  buyItem,
  shipItem,
  shipDone,
  complete,
} from './services.js';

import { getSession, saveSession } from './session.js';
import { initDB, getConnection, MySQLResultRows, MySQLResultSetHeader } from './db.js';

const execFile = promisify(exec);
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const app = new Hono();

// Initialize database
initDB();

// Middleware
app.use('*', async (c, next) => {
  c.header('Cache-Control', 'private');
  await next();
});

// Helper functions
async function getCurrentUser(c: Context): Promise<{ user: User | null; errCode: number; errMsg: string }> {
  const session = await getSession(c);

  if (!session.userId) {
    return { user: null, errCode: 404, errMsg: 'no session' };
  }

  const conn = await getConnection();
  try {
    const [rows] = await conn.query<MySQLResultRows>(
      'SELECT * FROM users WHERE id = ?',
      [session.userId]
    );

    if (rows.length === 0) {
      return { user: null, errCode: 404, errMsg: 'user not found' };
    }

    return { user: rows[0] as User, errCode: 0, errMsg: '' };
  } finally {
    conn.release();
  }
}

function outputErrorMsg(status: ContentfulStatusCode, msg: string): never {
  throw new HTTPException(status, { message: msg });
}

function httpError(status: number, msg: string): never {
  return outputErrorMsg(status as ContentfulStatusCode, msg);
}

// API Routes
app.post('/initialize', async (c) => {
  const body = await c.req.json<ReqInitialize>();

  try {
    await execFile('../init.sh');
  } catch (err) {
    console.error('Failed to execute init.sh:', err);
    httpError(500, 'exec init.sh error');
  }

  const conn = await getConnection();
  try {
    await conn.execute(
      "INSERT INTO `configs` (`name`, `val`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `val` = VALUES(`val`)",
      ['payment_service_url', body.payment_service_url]
    );

    await conn.execute(
      "INSERT INTO `configs` (`name`, `val`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `val` = VALUES(`val`)",
      ['shipment_service_url', body.shipment_service_url]
    );
  } finally {
    conn.release();
  }

  const res: ResInitialize = {
    campaign: 0,
    language: 'node',
  };

  return c.json(res);
});

app.get('/new_items.json', async (c) => {
  const itemIDStr = c.req.query('item_id');
  const createdAtStr = c.req.query('created_at');

  let itemID = 0;
  let createdAt = 0;

  if (itemIDStr && createdAtStr) {
    itemID = parseInt(itemIDStr, 10);
    createdAt = parseInt(createdAtStr, 10);
  }

  const conn = await getConnection();
  try {
    let items: Item[] = [];

    if (itemID > 0 && createdAt > 0) {
      const [rows] = await conn.query<MySQLResultRows>(
        'SELECT * FROM items WHERE status IN (?, ?) AND (created_at < ? OR (created_at = ? AND id < ?)) ORDER BY created_at DESC, id DESC LIMIT ?',
        [ITEM_STATUS.ON_SALE, ITEM_STATUS.SOLD_OUT, new Date(createdAt * 1000), new Date(createdAt * 1000), itemID, ITEMS_PER_PAGE + 1]
      );
      items = rows as Item[];
    } else {
      const [rows] = await conn.query<MySQLResultRows>(
        'SELECT * FROM items WHERE status IN (?, ?) ORDER BY created_at DESC, id DESC LIMIT ?',
        [ITEM_STATUS.ON_SALE, ITEM_STATUS.SOLD_OUT, ITEMS_PER_PAGE + 1]
      );
      items = rows as Item[];
    }

    const itemSimples: ItemSimple[] = [];
    for (const item of items.slice(0, ITEMS_PER_PAGE)) {
      const seller = await getUserSimpleByID(conn, item.seller_id);
      const category = await getCategoryByID(conn, item.category_id);

      itemSimples.push({
        id: item.id,
        seller_id: item.seller_id,
        seller: seller!,
        status: item.status,
        name: item.name,
        price: item.price,
        image_url: getImageURL(item.image_name),
        category_id: item.category_id,
        category: category!,
        created_at: Math.floor(item.created_at.getTime() / 1000),
      });
    }

    const hasNext = items.length > ITEMS_PER_PAGE;

    const res: ResNewItems = {
      has_next: hasNext,
      items: itemSimples,
    };

    return c.json(res);
  } finally {
    conn.release();
  }
});

app.get('/new_items/:root_category_id.json', async (c) => {
  const path = c.req.path;
  const match = path.match(/\/new_items\/([0-9]+)\.json/);
  if (!match) {
    return c.notFound();
  }
  const rootCategoryIDStr = match[1]!;
  const rootCategoryID = parseInt(rootCategoryIDStr, 10);

  const itemIDStr = c.req.query('item_id');
  const createdAtStr = c.req.query('created_at');

  let itemID = 0;
  let createdAt = 0;

  if (itemIDStr && createdAtStr) {
    itemID = parseInt(itemIDStr, 10);
    createdAt = parseInt(createdAtStr, 10);
  }

  const conn = await getConnection();
  try {
    const rootCategory = await getCategoryByID(conn, rootCategoryID);
    if (!rootCategory) {
      httpError(404, 'category not found');
    }

    const [categoryRows] = await conn.query<MySQLResultRows>(
      'SELECT id FROM categories WHERE parent_id = ?',
      [rootCategoryID]
    );

    const categoryIDs = [rootCategoryID];
    for (const row of categoryRows) {
      categoryIDs.push(row['id']);
    }

    let items: Item[] = [];

    if (itemID > 0 && createdAt > 0) {
      const placeholders = categoryIDs.map(() => '?').join(',');
      const [rows] = await conn.query<MySQLResultRows>(
        `SELECT * FROM items WHERE status IN (?, ?) AND category_id IN (${placeholders}) AND (created_at < ? OR (created_at = ? AND id < ?)) ORDER BY created_at DESC, id DESC LIMIT ?`,
        [ITEM_STATUS.ON_SALE, ITEM_STATUS.SOLD_OUT, ...categoryIDs, new Date(createdAt * 1000), new Date(createdAt * 1000), itemID, ITEMS_PER_PAGE + 1]
      );
      items = rows as Item[];
    } else {
      const placeholders = categoryIDs.map(() => '?').join(',');
      const [rows] = await conn.query<MySQLResultRows>(
        `SELECT * FROM items WHERE status IN (?, ?) AND category_id IN (${placeholders}) ORDER BY created_at DESC, id DESC LIMIT ?`,
        [ITEM_STATUS.ON_SALE, ITEM_STATUS.SOLD_OUT, ...categoryIDs, ITEMS_PER_PAGE + 1]
      );
      items = rows as Item[];
    }

    const itemSimples: ItemSimple[] = [];
    for (const item of items.slice(0, ITEMS_PER_PAGE)) {
      const seller = await getUserSimpleByID(conn, item.seller_id);
      const category = await getCategoryByID(conn, item.category_id);

      itemSimples.push({
        id: item.id,
        seller_id: item.seller_id,
        seller: seller!,
        status: item.status,
        name: item.name,
        price: item.price,
        image_url: getImageURL(item.image_name),
        category_id: item.category_id,
        category: category!,
        created_at: Math.floor(item.created_at.getTime() / 1000),
      });
    }

    const hasNext = items.length > ITEMS_PER_PAGE;

    const res: ResNewItems = {
      root_category_id: rootCategoryID,
      root_category_name: rootCategory.category_name,
      has_next: hasNext,
      items: itemSimples,
    };

    return c.json(res);
  } finally {
    conn.release();
  }
});

app.get('/users/transactions.json', async (c) => {
  const { user, errCode, errMsg } = await getCurrentUser(c);
  if (errCode !== 0) {
    outputErrorMsg(errCode as ContentfulStatusCode, errMsg);
  }

  const itemIDStr = c.req.query('item_id');
  const createdAtStr = c.req.query('created_at');

  let itemID = 0;
  let createdAt = 0;

  if (itemIDStr && createdAtStr) {
    itemID = parseInt(itemIDStr, 10);
    createdAt = parseInt(createdAtStr, 10);
  }

  const conn = await getConnection();
  try {
    await conn.beginTransaction();

    let items: Item[] = [];

    if (itemID > 0 && createdAt > 0) {
      const [rows] = await conn.query<MySQLResultRows>(
        'SELECT * FROM items WHERE (seller_id = ? OR buyer_id = ?) AND status IN (?, ?, ?, ?, ?) AND (created_at < ? OR (created_at = ? AND id < ?)) ORDER BY created_at DESC, id DESC LIMIT ?',
        [
          user!.id,
          user!.id,
          ITEM_STATUS.ON_SALE,
          ITEM_STATUS.TRADING,
          ITEM_STATUS.SOLD_OUT,
          ITEM_STATUS.CANCEL,
          ITEM_STATUS.STOP,
          new Date(createdAt * 1000),
          new Date(createdAt * 1000),
          itemID,
          TRANSACTIONS_PER_PAGE + 1,
        ]
      );
      items = rows as Item[];
    } else {
      const [rows] = await conn.query<MySQLResultRows>(
        'SELECT * FROM items WHERE (seller_id = ? OR buyer_id = ?) AND status IN (?, ?, ?, ?, ?) ORDER BY created_at DESC, id DESC LIMIT ?',
        [
          user!.id,
          user!.id,
          ITEM_STATUS.ON_SALE,
          ITEM_STATUS.TRADING,
          ITEM_STATUS.SOLD_OUT,
          ITEM_STATUS.CANCEL,
          ITEM_STATUS.STOP,
          TRANSACTIONS_PER_PAGE + 1,
        ]
      );
      items = rows as Item[];
    }

    const itemDetails: ItemDetail[] = [];

    for (const item of items.slice(0, TRANSACTIONS_PER_PAGE)) {
      const seller = await getUserSimpleByID(conn, item.seller_id);
      const category = await getCategoryByID(conn, item.category_id);

      const itemDetail: ItemDetail = {
        id: item.id,
        seller_id: item.seller_id,
        seller: seller!,
        status: item.status,
        name: item.name,
        price: item.price,
        description: item.description,
        image_url: getImageURL(item.image_name),
        category_id: item.category_id,
        category: category!,
        created_at: Math.floor(item.created_at.getTime() / 1000),
      };

      if (item.buyer_id !== 0) {
        const buyer = await getUserSimpleByID(conn, item.buyer_id);
        itemDetail.buyer_id = item.buyer_id;
        itemDetail.buyer = buyer!;
      }

      const [evidenceRows] = await conn.query<MySQLResultRows>(
        'SELECT * FROM transaction_evidences WHERE item_id = ?',
        [item.id]
      );

      if (evidenceRows.length > 0) {
        const transactionEvidence = evidenceRows[0] as TransactionEvidence;

        const [shippingRows] = await conn.query<MySQLResultRows>(
          'SELECT * FROM shippings WHERE transaction_evidence_id = ?',
          [transactionEvidence.id]
        );

        if (shippingRows.length > 0) {
          const shipping = shippingRows[0] as Shipping;

          const shipmentServiceURL = await getShipmentServiceURL(conn);
          const statusReq = {
            reserve_id: shipping.reserve_id,
          };

          let shippingStatus = shipping.status;
          try {
            const statusRes = await fetch(`${shipmentServiceURL}/status`, {
              method: 'GET',
              headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${SHIPMENT_SERVICE_ISUCARI_API_KEY}`,
              },
              body: JSON.stringify(statusReq),
            });

            if (statusRes.ok) {
              const shipmentStatus = await statusRes.json() as { status: string };
              shippingStatus = shipmentStatus.status;
            }
          } catch (err) {
            // If external service is down, use local status
          }

          itemDetail.transaction_evidence_id = transactionEvidence.id;
          itemDetail.transaction_evidence_status = transactionEvidence.status;
          itemDetail.shipping_status = shippingStatus;
        }
      }

      itemDetails.push(itemDetail);
    }

    await conn.commit();

    const hasNext = items.length > TRANSACTIONS_PER_PAGE;

    const res: ResTransactions = {
      has_next: hasNext,
      items: itemDetails,
    };

    return c.json(res);
  } catch (err) {
    await conn.rollback();
    throw err;
  } finally {
    conn.release();
  }
});

app.get('/users/:user_id.json', async (c) => {
  const path = c.req.path;
  const match = path.match(/\/users\/([0-9]+)\.json/);
  if (!match) {
    return c.notFound();
  }
  const userIDStr = match[1]!;
  const userID = parseInt(userIDStr, 10);

  const itemIDStr = c.req.query('item_id');
  const createdAtStr = c.req.query('created_at');

  let itemID = 0;
  let createdAt = 0;

  if (itemIDStr && createdAtStr) {
    itemID = parseInt(itemIDStr, 10);
    createdAt = parseInt(createdAtStr, 10);
  }

  const conn = await getConnection();
  try {
    const userSimple = await getUserSimpleByID(conn, userID);
    if (!userSimple) {
      httpError(404, 'user not found');
    }

    await conn.beginTransaction();

    let items: Item[] = [];

    if (itemID > 0 && createdAt > 0) {
      const [rows] = await conn.query<MySQLResultRows>(
        'SELECT * FROM items WHERE seller_id = ? AND status IN (?, ?, ?) AND (created_at < ? OR (created_at = ? AND id < ?)) ORDER BY created_at DESC, id DESC LIMIT ?',
        [userID, ITEM_STATUS.ON_SALE, ITEM_STATUS.TRADING, ITEM_STATUS.SOLD_OUT, new Date(createdAt * 1000), new Date(createdAt * 1000), itemID, ITEMS_PER_PAGE + 1]
      );
      items = rows as Item[];
    } else {
      const [rows] = await conn.query<MySQLResultRows>(
        'SELECT * FROM items WHERE seller_id = ? AND status IN (?, ?, ?) ORDER BY created_at DESC, id DESC LIMIT ?',
        [userID, ITEM_STATUS.ON_SALE, ITEM_STATUS.TRADING, ITEM_STATUS.SOLD_OUT, ITEMS_PER_PAGE + 1]
      );
      items = rows as Item[];
    }

    const itemSimples: ItemSimple[] = [];

    for (const item of items.slice(0, ITEMS_PER_PAGE)) {
      const category = await getCategoryByID(conn, item.category_id);

      itemSimples.push({
        id: item.id,
        seller_id: item.seller_id,
        seller: userSimple!,
        status: item.status,
        name: item.name,
        price: item.price,
        image_url: getImageURL(item.image_name),
        category_id: item.category_id,
        category: category!,
        created_at: Math.floor(item.created_at.getTime() / 1000),
      });
    }

    await conn.commit();

    const hasNext = items.length > ITEMS_PER_PAGE;

    const res: ResUserItems = {
      user: userSimple!,
      has_next: hasNext,
      items: itemSimples,
    };

    return c.json(res);
  } catch (err) {
    await conn.rollback();
    throw err;
  } finally {
    conn.release();
  }
});

app.get('/items/:item_id.json', async (c) => {
  const path = c.req.path;
  const match = path.match(/\/items\/([0-9]+)\.json/);
  if (!match) {
    return c.notFound();
  }
  const itemIDStr = match[1]!;
  const itemID = parseInt(itemIDStr, 10);

  const { user } = await getCurrentUser(c);

  const conn = await getConnection();
  try {
    const [itemRows] = await conn.query<MySQLResultRows>(
      'SELECT * FROM items WHERE id = ?',
      [itemID]
    );

    if (itemRows.length === 0) {
      httpError(404, 'item not found');
    }

    const item = itemRows[0] as Item;

    const seller = await getUserSimpleByID(conn, item.seller_id);
    const category = await getCategoryByID(conn, item.category_id);

    const itemDetail: ItemDetail = {
      id: item.id,
      seller_id: item.seller_id,
      seller: seller!,
      status: item.status,
      name: item.name,
      price: item.price,
      description: item.description,
      image_url: getImageURL(item.image_name),
      category_id: item.category_id,
      category: category!,
      created_at: Math.floor(item.created_at.getTime() / 1000),
    };

    if (user && (user.id === item.seller_id || user.id === item.buyer_id)) {
      if (item.buyer_id !== 0) {
        const buyer = await getUserSimpleByID(conn, item.buyer_id);
        itemDetail.buyer_id = item.buyer_id;
        itemDetail.buyer = buyer!;
      }

      const [evidenceRows] = await conn.query<MySQLResultRows>(
        'SELECT * FROM transaction_evidences WHERE item_id = ?',
        [item.id]
      );

      if (evidenceRows.length > 0) {
        const transactionEvidence = evidenceRows[0] as TransactionEvidence;

        const [shippingRows] = await conn.query<MySQLResultRows>(
          'SELECT * FROM shippings WHERE transaction_evidence_id = ?',
          [transactionEvidence.id]
        );

        if (shippingRows.length > 0) {
          const shipping = shippingRows[0] as Shipping;

          const shipmentServiceURL = await getShipmentServiceURL(conn);
          const statusReq = {
            reserve_id: shipping.reserve_id,
          };

          let shippingStatus = shipping.status;
          try {
            const statusRes = await fetch(`${shipmentServiceURL}/status`, {
              method: 'GET',
              headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${SHIPMENT_SERVICE_ISUCARI_API_KEY}`,
              },
              body: JSON.stringify(statusReq),
            });

            if (statusRes.ok) {
              const shipmentStatus = await statusRes.json() as { status: string };
              shippingStatus = shipmentStatus.status;
            }
          } catch (err) {
            // If external service is down, use local status
          }

          itemDetail.transaction_evidence_id = transactionEvidence.id;
          itemDetail.transaction_evidence_status = transactionEvidence.status;
          itemDetail.shipping_status = shippingStatus;
        }
      }
    }

    return c.json(itemDetail);
  } finally {
    conn.release();
  }
});

app.post('/items/edit', async (c) => {
  const body = await c.req.json<ReqItemEdit>();

  const session = await getSession(c);
  if (session.csrfToken !== body.csrf_token) {
    httpError(422, 'csrf token error');
  }

  const { user, errCode, errMsg } = await getCurrentUser(c);
  if (errCode !== 0) {
    outputErrorMsg(errCode as ContentfulStatusCode, errMsg);
  }

  if (body.item_price < ITEM_MIN_PRICE || body.item_price > ITEM_MAX_PRICE) {
    httpError(400, ITEM_PRICE_ERR_MSG);
  }

  const conn = await getConnection();
  try {
    await conn.beginTransaction();

    const [items] = await conn.query<MySQLResultRows>(
      'SELECT * FROM items WHERE id = ? FOR UPDATE',
      [body.item_id]
    );

    if (items.length === 0) {
      await conn.rollback();
      httpError(404, 'item not found');
    }

    const item = items[0] as Item;

    if (item.seller_id !== user!.id) {
      await conn.rollback();
      httpError(403, '自分の商品以外は編集できません');
    }

    if (item.status !== ITEM_STATUS.ON_SALE) {
      await conn.rollback();
      httpError(403, '販売中の商品以外編集できません');
    }

    const now = new Date();

    await conn.execute(
      'UPDATE items SET price = ?, updated_at = ? WHERE id = ?',
      [body.item_price, now, item.id]
    );

    const [updatedItems] = await conn.query<MySQLResultRows>(
      'SELECT * FROM items WHERE id = ?',
      [item.id]
    );

    const updatedItem = updatedItems[0] as Item;

    await conn.commit();

    const res: ResItemEdit = {
      item_id: updatedItem.id,
      item_price: updatedItem.price,
      item_created_at: Math.floor(updatedItem.created_at.getTime() / 1000),
      item_updated_at: Math.floor(updatedItem.updated_at.getTime() / 1000),
    };

    return c.json(res);
  } catch (err) {
    await conn.rollback();
    throw err;
  } finally {
    conn.release();
  }
});

app.post('/buy', async (c) => {
  const body = await c.req.json<ReqBuy>();

  const session = await getSession(c);
  if (session.csrfToken !== body.csrf_token) {
    httpError(422, 'csrf token error');
  }

  const { user, errCode, errMsg } = await getCurrentUser(c);
  if (errCode !== 0) {
    outputErrorMsg(errCode as ContentfulStatusCode, errMsg);
  }

  const conn = await getConnection();
  try {
    const result = await buyItem(conn, user!.id, body.item_id, body.token);

    if ('error' in result) {
      if (result.error === 'item not found' || result.error === 'buyer not found' || result.error === 'seller not found') {
        httpError(404, result.error);
      } else if (result.error === 'item is not for sale' || result.error === 'cannot buy your own item') {
        httpError(403, result.error);
      } else if (result.error === 'カード情報に誤りがあります' || result.error === 'カードの残高が足りません' || result.error === '想定外のエラー') {
        httpError(400, result.error);
      } else {
        httpError(500, result.error);
      }
      return;
    }

    const res: ResTransactionEvidence = {
      transaction_evidence_id: result.transactionEvidenceID,
    };

    return c.json(res);
  } finally {
    conn.release();
  }
});

app.post('/sell', async (c) => {
  const formData = await c.req.formData();
  const body: ReqSell = {
    csrf_token: formData.get('csrf_token') as string,
    name: formData.get('name') as string,
    description: formData.get('description') as string,
    price: parseInt(formData.get('price') as string, 10),
    category_id: parseInt(formData.get('category_id') as string, 10),
  };
  const imageFile = formData.get('image') as File | null;

  const session = await getSession(c);
  if (session.csrfToken !== body.csrf_token) {
    httpError(422, 'csrf token error');
  }

  const { user, errCode, errMsg } = await getCurrentUser(c);
  if (errCode !== 0) {
    outputErrorMsg(errCode as ContentfulStatusCode, errMsg);
  }

  if (body.name === '' || body.description === '' || body.price === 0 || body.category_id === 0) {
    httpError(400, 'all parameters are required');
  }

  if (body.price < ITEM_MIN_PRICE || body.price > ITEM_MAX_PRICE) {
    httpError(400, ITEM_PRICE_ERR_MSG);
  }

  const conn = await getConnection();
  try {
    const category = await getCategoryByID(conn, body.category_id);
    if (!category) {
      httpError(400, 'category id error');
      return;
    }

    if (!imageFile) {
      httpError(400, 'image required');
      return;
    }

    const ext = extname(imageFile.name);
    const imageName = secureRandomStr(16) + ext;

    const uploadPath = join(__dirname, '../../public/upload', imageName);
    const buffer = Buffer.from(await imageFile.arrayBuffer());
    await writeFile(uploadPath, buffer);

    await conn.beginTransaction();

    const [sellers] = await conn.query<MySQLResultRows>(
      'SELECT * FROM users WHERE id = ? FOR UPDATE',
      [user!.id]
    );

    if (sellers.length === 0) {
      await conn.rollback();
      httpError(404, 'user not found');
    }

    const seller = sellers[0] as User;

    const now = new Date();

    const [result] = await conn.execute(
      'INSERT INTO items (seller_id, status, name, price, description, image_name, category_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)',
      [seller.id, ITEM_STATUS.ON_SALE, body.name, body.price, body.description, imageName, body.category_id, now, now]
    );

    const itemID = (result as MySQLResultSetHeader).insertId;

    await conn.execute(
      'UPDATE users SET num_sell_items = ?, last_bump = ? WHERE id = ?',
      [seller.num_sell_items + 1, now, seller.id]
    );

    await conn.commit();

    const res: ResSell = {
      id: itemID,
    };

    return c.json(res);
  } catch (err) {
    await conn.rollback();
    throw err;
  } finally {
    conn.release();
  }
});

app.post('/ship', async (c) => {
  const body = await c.req.json<ReqShip>();

  const session = await getSession(c);
  if (session.csrfToken !== body.csrf_token) {
    httpError(422, 'csrf token error');
  }

  const { user, errCode, errMsg } = await getCurrentUser(c);
  if (errCode !== 0) {
    outputErrorMsg(errCode as ContentfulStatusCode, errMsg);
  }

  const conn = await getConnection();
  try {
    const result = await shipItem(conn, user!.id, body.item_id);

    if ('error' in result) {
      if (result.error === 'item not found' || result.error === 'transaction evidence not found' ||
          result.error === 'shipping not found' || result.error === 'buyer not found' ||
          result.error === 'seller not found') {
        httpError(404, result.error);
      } else if (result.error === 'forbidden') {
        httpError(403, result.error);
      } else if (result.error === 'item is not waiting for shipping') {
        httpError(400, result.error);
      } else {
        httpError(500, result.error);
      }
      return;
    }

    const res: ResShip = {
      path: result.path,
      reserve_id: result.reserveID,
    };

    return c.json(res);
  } finally {
    conn.release();
  }
});

app.post('/ship_done', async (c) => {
  const body = await c.req.json<ReqShipDone>();

  const session = await getSession(c);
  if (session.csrfToken !== body.csrf_token) {
    httpError(422, 'csrf token error');
  }

  const { user, errCode, errMsg } = await getCurrentUser(c);
  if (errCode !== 0) {
    outputErrorMsg(errCode as ContentfulStatusCode, errMsg);
  }

  const conn = await getConnection();
  try {
    const result = await shipDone(conn, user!.id, body.item_id);

    if ('error' in result) {
      if (result.error === 'item not found' || result.error === 'transaction evidence not found' ||
          result.error === 'shipping not found') {
        httpError(404, result.error);
      } else if (result.error === 'forbidden' ||
                 result.error === 'shipment service側で配送中か配送完了になっていません') {
        httpError(403, result.error);
      } else if (result.error === 'item is not waiting for shipping' ||
                 result.error === 'shipping not requested' ||
                 result.error === 'item is still shipping' ||
                 result.error === 'item has already arrived') {
        httpError(400, result.error);
      } else {
        httpError(500, result.error);
      }
      return;
    }

    const res: ResTransactionEvidence = {
      transaction_evidence_id: result.transactionEvidenceID,
    };

    return c.json(res);
  } finally {
    conn.release();
  }
});

app.post('/complete', async (c) => {
  const body = await c.req.json<ReqComplete>();

  const session = await getSession(c);
  if (session.csrfToken !== body.csrf_token) {
    httpError(422, 'csrf token error');
  }

  const { user, errCode, errMsg } = await getCurrentUser(c);
  if (errCode !== 0) {
    outputErrorMsg(errCode as ContentfulStatusCode, errMsg);
  }

  const conn = await getConnection();
  try {
    const result = await complete(conn, user!.id, body.item_id);

    if ('error' in result) {
      if (result.error === 'item not found' || result.error === 'transaction evidence not found' ||
          result.error === 'shipping not found') {
        httpError(404, result.error);
      } else if (result.error === 'forbidden') {
        httpError(403, result.error);
      } else if (result.error === 'item is not waiting for completion' ||
                 result.error === 'shipment service側で配送完了になっていません') {
        httpError(400, result.error);
      } else {
        httpError(500, result.error);
      }
      return;
    }

    const res: ResTransactionEvidence = {
      transaction_evidence_id: result.transactionEvidenceID,
    };

    return c.json(res);
  } finally {
    conn.release();
  }
});

app.get('/transactions/:transaction_evidence_id.png', async (c) => {
  const path = c.req.path;
  const match = path.match(/\/transactions\/([0-9]+)\.png/);
  if (!match) {
    httpError(400, 'incorrect transaction_evidence id');
  }

  const transactionEvidenceIDStr = match![1]!;
  const transactionEvidenceID = parseInt(transactionEvidenceIDStr, 10);

  if (isNaN(transactionEvidenceID) || transactionEvidenceID <= 0) {
    httpError(400, 'incorrect transaction_evidence id');
  }

  const { user, errCode, errMsg } = await getCurrentUser(c);
  if (errCode !== 0) {
    httpError(403, errMsg);
  }

  const conn = await getConnection();
  try {
    const [evidences] = await conn.query<MySQLResultRows>(
      'SELECT * FROM transaction_evidences WHERE id = ?',
      [transactionEvidenceID]
    );

    if (evidences.length === 0) {
      httpError(404, 'transaction_evidences not found');
    }

    const transactionEvidence = evidences[0] as TransactionEvidence;

    if (transactionEvidence.seller_id !== user!.id) {
      httpError(403, '権限がありません');
    }

    const [shippings] = await conn.query<MySQLResultRows>(
      'SELECT * FROM shippings WHERE transaction_evidence_id = ?',
      [transactionEvidence.id]
    );

    if (shippings.length === 0) {
      httpError(404, 'shippings not found');
    }

    const shipping = shippings[0] as Shipping;

    if (shipping.status !== SHIPPINGS_STATUS.WAIT_PICKUP && shipping.status !== SHIPPINGS_STATUS.SHIPPING) {
      httpError(403, 'qrcode not available');
    }

    if (shipping.img_binary.length === 0) {
      httpError(500, 'empty qrcode image');
    }

    c.header('Content-Type', 'image/png');
    return c.body(new Uint8Array(shipping.img_binary));
  } finally {
    conn.release();
  }
});

app.post('/bump', async (c) => {
  const body = await c.req.json<ReqBump>();

  const session = await getSession(c);
  if (session.csrfToken !== body.csrf_token) {
    httpError(422, 'csrf token error');
  }

  const { user, errCode, errMsg } = await getCurrentUser(c);
  if (errCode !== 0) {
    outputErrorMsg(errCode as ContentfulStatusCode, errMsg);
  }

  const conn = await getConnection();
  try {
    await conn.beginTransaction();

    const [items] = await conn.query<MySQLResultRows>(
      'SELECT * FROM items WHERE id = ? FOR UPDATE',
      [body.item_id]
    );

    if (items.length === 0) {
      await conn.rollback();
      httpError(404, 'item not found');
    }

    const item = items[0] as Item;

    if (item.seller_id !== user!.id) {
      await conn.rollback();
      httpError(403, '自分の商品以外は編集できません');
    }

    const [sellers] = await conn.query<MySQLResultRows>(
      'SELECT * FROM users WHERE id = ? FOR UPDATE',
      [user!.id]
    );

    if (sellers.length === 0) {
      await conn.rollback();
      httpError(404, 'user not found');
    }

    const seller = sellers[0] as User;

    const now = new Date();

    if (seller.last_bump.getTime() + BUMP_CHARGE_SECONDS > now.getTime()) {
      await conn.rollback();
      httpError(403, 'Bump not allowed');
    }

    await conn.execute(
      'UPDATE items SET created_at = ?, updated_at = ? WHERE id = ?',
      [now, now, item.id]
    );

    await conn.execute(
      'UPDATE users SET last_bump = ? WHERE id = ?',
      [now, seller.id]
    );

    const [newItems] = await conn.query<MySQLResultRows>(
      'SELECT * FROM items WHERE id = ?',
      [item.id]
    );

    const newItem = newItems[0] as Item;

    await conn.commit();

    const res: ResBump = {
      item_id: newItem.id,
      item_price: newItem.price,
      item_created_at: Math.floor(newItem.created_at.getTime() / 1000),
      item_updated_at: Math.floor(newItem.updated_at.getTime() / 1000),
    };

    return c.json(res);
  } catch (err) {
    await conn.rollback();
    throw err;
  } finally {
    conn.release();
  }
});

app.get('/settings', async (c) => {
  const session = await getSession(c);
  const csrfToken = session.csrfToken || secureRandomStr(20);

  const { user } = await getCurrentUser(c);

  const conn = await getConnection();
  try {
    const [categories] = await conn.query<MySQLResultRows>(
      'SELECT * FROM categories'
    );

    const paymentServiceURL = await getPaymentServiceURL(conn);

    let res: ResSettings;

    if (user) {
      res = {
        csrf_token: csrfToken,
        user: user,
        categories: categories as Category[],
        payment_service_url: paymentServiceURL,
      };
    } else {
      res = {
        csrf_token: csrfToken,
        user: null,
        categories: categories as Category[],
        payment_service_url: paymentServiceURL,
      };
    }

    if (!session.csrfToken) {
      await saveSession(c, { ...session, csrfToken });
    }

    return c.json(res);
  } finally {
    conn.release();
  }
});

app.post('/login', async (c) => {
  const body = await c.req.json<ReqLogin>();

  const conn = await getConnection();
  try {
    const [users] = await conn.query<MySQLResultRows>(
      'SELECT * FROM users WHERE account_name = ?',
      [body.account_name]
    );

    if (users.length === 0) {
      httpError(401, 'アカウント名かパスワードが間違えています');
    }

    const user = users[0] as User;

    const passwordMatch = await verifyPassword(body.password, user.hashed_password);
    if (!passwordMatch) {
      httpError(401, 'アカウント名かパスワードが間違えています');
    }

    await saveSession(c, {
      userId: user.id,
      csrfToken: secureRandomStr(20),
    });

    const res: ResLogin = {
      id: user.id,
      account_name: user.account_name,
      address: user.address,
    };

    return c.json(res);
  } finally {
    conn.release();
  }
});

app.post('/register', async (c) => {
  const body = await c.req.json<ReqRegister>();

  if (body.account_name === '' || body.password === '' || body.address === '') {
    httpError(400, 'all parameters are required');
  }

  const hashedPassword = await hashPassword(body.password);

  const conn = await getConnection();
  try {
    const [result] = await conn.execute(
      'INSERT INTO users (account_name, hashed_password, address) VALUES (?, ?, ?)',
      [body.account_name, hashedPassword, body.address]
    );

    const userID = (result as MySQLResultSetHeader).insertId;

    await saveSession(c, {
      userId: userID,
      csrfToken: secureRandomStr(20),
    });

    const res: ResRegister = {
      id: userID,
      account_name: body.account_name,
      address: body.address,
    };

    return c.json(res);
  } finally {
    conn.release();
  }
});

app.get('/reports.json', async (c) => {
  const conn = await getConnection();
  try {
    const [transactionEvidences] = await conn.query<MySQLResultRows>(
      'SELECT * FROM transaction_evidences WHERE id > 15007'
    );

    return c.json(transactionEvidences);
  } finally {
    conn.release();
  }
});

// Frontend routes
const frontendRoutes = [
  '/',
  '/login',
  '/register',
  '/timeline',
  '/categories/:category_id/items',
  '/sell',
  '/items/:item_id',
  '/items/:item_id/edit',
  '/items/:item_id/buy',
  '/buy/complete',
  '/transactions/:transaction_id',
  '/users/:user_id',
  '/settings',
];

frontendRoutes.forEach((route) => {
  app.get(route, async (c) => {
    const indexHtml = await readFile(join(__dirname, '../../public/index.html'), 'utf-8');
    return c.html(indexHtml);
  });
});

// Static files
app.use('/static/*', serveStatic({ root: join(__dirname, '../../public') }));
app.use('/upload/*', serveStatic({ root: join(__dirname, '../../public') }));

// Error handler
app.onError((err, c) => {
  if (err instanceof HTTPException) {
    return c.json({ error: err.message }, err.status);
  }
  console.error(err);
  return c.json({ error: 'internal server error' }, 500);
});

serve({
  fetch: app.fetch,
  port: 8000,
  hostname: '0.0.0.0',
});
