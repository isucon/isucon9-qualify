import { PoolConnection } from 'mysql2/promise';
import {
  User,
  UserSimple,
  Item,
  Category,
  TransactionEvidence
} from './types.js';
import {
  DEFAULT_PAYMENT_SERVICE_URL,
  DEFAULT_SHIPMENT_SERVICE_URL,
  PAYMENT_SERVICE_ISUCARI_API_KEY,
  PAYMENT_SERVICE_ISUCARI_SHOP_ID,
  SHIPMENT_SERVICE_ISUCARI_API_KEY,
} from './constants.js';
import { MySQLResultRows, MySQLResultSetHeader } from './db.js';

export async function getUserSimpleByID(conn: PoolConnection, userID: number): Promise<UserSimple | null> {
  const [rows] = await conn.query<MySQLResultRows>(
    'SELECT id, account_name, num_sell_items FROM users WHERE id = ?',
    [userID]
  );

  if (rows.length === 0) {
    return null;
  }

  const row = rows[0]!;
  return {
    id: row['id'],
    account_name: row['account_name'],
    num_sell_items: row['num_sell_items'],
  };
}

export async function getCategoryByID(conn: PoolConnection, categoryID: number): Promise<Category | null> {
  const [rows] = await conn.query<MySQLResultRows>(
    'SELECT * FROM categories WHERE id = ?',
    [categoryID]
  );

  if (rows.length === 0) {
    return null;
  }

  const category = rows[0]!;

  if (category['parent_id'] !== 0) {
    const parent = await getCategoryByID(conn, category['parent_id']);
    if (parent) {
      category['parent_category_name'] = parent.category_name;
    }
  }

  return {
    id: category['id'],
    parent_id: category['parent_id'],
    category_name: category['category_name'],
    parent_category_name: category['parent_category_name'] || '',
  };
}

export async function getConfigByName(conn: PoolConnection, name: string): Promise<string | null> {
  const [rows] = await conn.query<MySQLResultRows>(
    'SELECT val FROM configs WHERE name = ?',
    [name]
  );

  if (rows.length === 0) {
    return null;
  }

  return rows[0]!['val'];
}

export async function getPaymentServiceURL(conn: PoolConnection): Promise<string> {
  const url = await getConfigByName(conn, 'payment_service_url');
  return url || DEFAULT_PAYMENT_SERVICE_URL;
}

export async function getShipmentServiceURL(conn: PoolConnection): Promise<string> {
  const url = await getConfigByName(conn, 'shipment_service_url');
  return url || DEFAULT_SHIPMENT_SERVICE_URL;
}

interface PaymentTokenReq {
  shop_id: string;
  token: string;
  api_key: string;
  price: number;
}

interface PaymentTokenRes {
  status: string;
}

export async function buyItem(
  conn: PoolConnection,
  buyerID: number,
  itemID: number,
  token: string
): Promise<{ transactionEvidenceID: number } | { error: string }> {
  await conn.beginTransaction();

  try {
    const [items] = await conn.query<MySQLResultRows>(
      'SELECT * FROM items WHERE id = ? FOR UPDATE',
      [itemID]
    );

    if (items.length === 0) {
      await conn.rollback();
      return { error: 'item not found' };
    }

    const item = items[0] as Item;

    if (item.status !== 'on_sale') {
      await conn.rollback();
      return { error: 'item is not for sale' };
    }

    if (item.seller_id === buyerID) {
      await conn.rollback();
      return { error: 'cannot buy your own item' };
    }

    const [buyers] = await conn.query<MySQLResultRows>(
      'SELECT * FROM users WHERE id = ? FOR UPDATE',
      [buyerID]
    );

    if (buyers.length === 0) {
      await conn.rollback();
      return { error: 'buyer not found' };
    }

    const buyer = buyers[0] as User;

    const now = new Date();

    await conn.execute(
      'UPDATE items SET status = ?, buyer_id = ?, updated_at = ? WHERE id = ?',
      ['trading', buyerID, now, itemID]
    );

    const [sellers] = await conn.query<MySQLResultRows>(
      'SELECT * FROM users WHERE id = ? FOR UPDATE',
      [item.seller_id]
    );

    if (sellers.length === 0) {
      await conn.rollback();
      return { error: 'seller not found' };
    }

    const seller = sellers[0] as User;

    const category = await getCategoryByID(conn, item.category_id);
    if (!category) {
      await conn.rollback();
      return { error: 'category not found' };
    }

    const [result] = await conn.execute(
      `INSERT INTO transaction_evidences
        (seller_id, buyer_id, status, item_id, item_name, item_price,
         item_description, item_category_id, item_root_category_id)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      [
        item.seller_id,
        buyerID,
        'wait_shipping',
        itemID,
        item.name,
        item.price,
        item.description,
        category.id,
        category.parent_id,
      ]
    );

    const transactionEvidenceID = (result as MySQLResultSetHeader).insertId;

    // Create shipment first (same order as Go implementation)
    const shipmentServiceURL = await getShipmentServiceURL(conn);
    const shipmentReq: ShipmentCreateReq = {
      to_address: buyer.address,
      to_name: buyer.account_name,
      from_address: seller.address,
      from_name: seller.account_name,
    };

    let shipmentRes;
    try {
      shipmentRes = await fetch(`${shipmentServiceURL}/create`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${SHIPMENT_SERVICE_ISUCARI_API_KEY}`,
        },
        body: JSON.stringify(shipmentReq),
      });
    } catch (err) {
      await conn.rollback();
      return { error: 'failed to request to shipment service' };
    }

    if (!shipmentRes.ok) {
      await conn.rollback();
      return { error: 'failed to request to shipment service' };
    }

    const shipmentCreate = await shipmentRes.json() as ShipmentCreateRes;

    // Then process payment
    const paymentServiceURL = await getPaymentServiceURL(conn);
    const paymentURL = `${paymentServiceURL}/token`;

    const paymentReq: PaymentTokenReq = {
      shop_id: PAYMENT_SERVICE_ISUCARI_SHOP_ID,
      token: token,
      api_key: PAYMENT_SERVICE_ISUCARI_API_KEY,
      price: item.price,
    };

    let response;
    try {
      response = await fetch(paymentURL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(paymentReq),
      });
    } catch (err) {
      await conn.rollback();
      return { error: 'payment service is failed' };
    }

    if (!response.ok) {
      await conn.rollback();
      return { error: 'payment service is failed' };
    }

    const paymentRes = await response.json() as PaymentTokenRes;

    if (paymentRes.status === 'invalid') {
      await conn.rollback();
      return { error: 'カード情報に誤りがあります' };
    }

    if (paymentRes.status === 'fail') {
      await conn.rollback();
      return { error: 'カードの残高が足りません' };
    }

    if (paymentRes.status !== 'ok') {
      await conn.rollback();
      return { error: '想定外のエラー' };
    }

    // Insert shipping record
    await conn.execute(
      `INSERT INTO shippings
        (transaction_evidence_id, status, item_name, item_id, reserve_id, reserve_time,
         to_address, to_name, from_address, from_name, img_binary)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      [
        transactionEvidenceID,
        'initial',
        item.name,
        itemID,
        shipmentCreate.reserve_id,
        shipmentCreate.reserve_time,
        buyer.address,
        buyer.account_name,
        seller.address,
        seller.account_name,
        Buffer.from(''),
      ]
    )

    await conn.commit();
    return { transactionEvidenceID };

  } catch (err) {
    await conn.rollback();
    throw err;
  }
}

interface ShipmentCreateReq {
  to_address: string;
  to_name: string;
  from_address: string;
  from_name: string;
}

interface ShipmentCreateRes {
  reserve_id: string;
  reserve_time: number;
}

interface ShipmentRequestReq {
  reserve_id: string;
}

interface ShipmentStatusReq {
  reserve_id: string;
}

interface ShipmentStatusRes {
  status: string;
  reserve_time: number;
}

export async function shipItem(
  conn: PoolConnection,
  sellerID: number,
  itemID: number
): Promise<{ path: string; reserveID: string } | { error: string }> {
  await conn.beginTransaction();

  try {
    const [items] = await conn.query<MySQLResultRows>(
      'SELECT * FROM items WHERE id = ? FOR UPDATE',
      [itemID]
    );

    if (items.length === 0) {
      await conn.rollback();
      return { error: 'item not found' };
    }

    const item = items[0] as Item;

    if (item.seller_id !== sellerID) {
      await conn.rollback();
      return { error: 'forbidden' };
    }

    const [evidences] = await conn.query<MySQLResultRows>(
      'SELECT * FROM transaction_evidences WHERE item_id = ? FOR UPDATE',
      [itemID]
    );

    if (evidences.length === 0) {
      await conn.rollback();
      return { error: 'transaction evidence not found' };
    }

    const transactionEvidence = evidences[0] as TransactionEvidence;

    if (transactionEvidence.status !== 'wait_shipping') {
      await conn.rollback();
      return { error: 'item is not waiting for shipping' };
    }

    const [shippings] = await conn.query<MySQLResultRows>(
      'SELECT * FROM shippings WHERE transaction_evidence_id = ? FOR UPDATE',
      [transactionEvidence.id]
    );

    if (shippings.length === 0) {
      await conn.rollback();
      return { error: 'shipping not found' };
    }

    // const shipping = shippings[0]!;

    const [buyers] = await conn.query<MySQLResultRows>(
      'SELECT * FROM users WHERE id = ?',
      [item.buyer_id]
    );

    if (buyers.length === 0) {
      await conn.rollback();
      return { error: 'buyer not found' };
    }

    const buyer = buyers[0] as User;

    const [sellers] = await conn.query<MySQLResultRows>(
      'SELECT * FROM users WHERE id = ?',
      [item.seller_id]
    );

    if (sellers.length === 0) {
      await conn.rollback();
      return { error: 'seller not found' };
    }

    const seller = sellers[0] as User;

    const shipmentServiceURL = await getShipmentServiceURL(conn);

    const createReq: ShipmentCreateReq = {
      to_address: buyer.address,
      to_name: buyer.account_name,
      from_address: seller.address,
      from_name: seller.account_name,
    };

    let createRes;
    try {
      createRes = await fetch(`${shipmentServiceURL}/create`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${SHIPMENT_SERVICE_ISUCARI_API_KEY}`,
        },
        body: JSON.stringify(createReq),
      });
    } catch (err) {
      await conn.rollback();
      return { error: 'failed to create shipment' };
    }

    if (!createRes.ok) {
      await conn.rollback();
      return { error: 'failed to create shipment' };
    }

    const shipmentCreate = await createRes.json() as ShipmentCreateRes;

    let requestRes;
    try {
      requestRes = await fetch(`${shipmentServiceURL}/request`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${SHIPMENT_SERVICE_ISUCARI_API_KEY}`,
        },
        body: JSON.stringify({ reserve_id: shipmentCreate.reserve_id }),
      });
    } catch (err) {
      await conn.rollback();
      return { error: 'failed to request shipment' };
    }

    if (!requestRes.ok) {
      await conn.rollback();
      return { error: 'failed to request shipment' };
    }

    const imgBinary = Buffer.from(await requestRes.arrayBuffer());

    const now = new Date();

    await conn.execute(
      'UPDATE shippings SET status = ?, img_binary = ?, reserve_id = ?, reserve_time = ?, updated_at = ? WHERE transaction_evidence_id = ?',
      ['wait_pickup', imgBinary, shipmentCreate.reserve_id, shipmentCreate.reserve_time, now, transactionEvidence.id]
    );

    await conn.commit();

    return {
      path: `/transactions/${transactionEvidence.id}.png`,
      reserveID: shipmentCreate.reserve_id,
    };

  } catch (err) {
    await conn.rollback();
    throw err;
  }
}

export async function shipDone(
  conn: PoolConnection,
  sellerID: number,
  itemID: number
): Promise<{ transactionEvidenceID: number } | { error: string }> {
  await conn.beginTransaction();

  try {
    const [items] = await conn.query<MySQLResultRows>(
      'SELECT * FROM items WHERE id = ? FOR UPDATE',
      [itemID]
    );

    if (items.length === 0) {
      await conn.rollback();
      return { error: 'item not found' };
    }

    const item = items[0] as Item;

    if (item.seller_id !== sellerID) {
      await conn.rollback();
      return { error: 'forbidden' };
    }

    const [evidences] = await conn.query<MySQLResultRows>(
      'SELECT * FROM transaction_evidences WHERE item_id = ? FOR UPDATE',
      [itemID]
    );

    if (evidences.length === 0) {
      await conn.rollback();
      return { error: 'transaction evidence not found' };
    }

    const transactionEvidence = evidences[0] as TransactionEvidence;

    if (transactionEvidence.status !== 'wait_shipping') {
      await conn.rollback();
      return { error: 'item is not waiting for shipping' };
    }

    const [shippings] = await conn.query<MySQLResultRows>(
      'SELECT * FROM shippings WHERE transaction_evidence_id = ? FOR UPDATE',
      [transactionEvidence.id]
    );

    if (shippings.length === 0) {
      await conn.rollback();
      return { error: 'shipping not found' };
    }

    const shipping = shippings[0]!;

    const now = new Date();

    switch (shipping['status']) {
      case 'initial':
        await conn.rollback();
        return { error: 'shipping not requested' };

      case 'wait_pickup':
        const shipmentServiceURL = await getShipmentServiceURL(conn);

        // First check the shipment status
        const statusCheckReq: ShipmentStatusReq = {
          reserve_id: shipping['reserve_id'],
        };

        let statusCheckRes;
        try {
          statusCheckRes = await fetch(`${shipmentServiceURL}/status`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${SHIPMENT_SERVICE_ISUCARI_API_KEY}`,
            },
            body: JSON.stringify(statusCheckReq),
          });
        } catch (err) {
          await conn.rollback();
          return { error: 'failed to get shipment status' };
        }

        if (!statusCheckRes.ok) {
          await conn.rollback();
          return { error: 'failed to get shipment status' };
        }

        const shipmentStatusCheck = await statusCheckRes.json() as ShipmentStatusRes;

        if (shipmentStatusCheck.status !== 'shipping' && shipmentStatusCheck.status !== 'done') {
          await conn.rollback();
          return { error: 'shipment service側で配送中か配送完了になっていません' };
        }

        const requestReq: ShipmentRequestReq = {
          reserve_id: shipping['reserve_id'],
        };

        let requestRes;
        try {
          requestRes = await fetch(`${shipmentServiceURL}/request`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${SHIPMENT_SERVICE_ISUCARI_API_KEY}`,
            },
            body: JSON.stringify(requestReq),
          });
        } catch (err) {
          await conn.rollback();
          return { error: 'failed to request shipment' };
        }

        if (!requestRes.ok) {
          await conn.rollback();
          return { error: 'failed to request shipment' };
        }

        await conn.execute(
          'UPDATE shippings SET status = ?, updated_at = ? WHERE transaction_evidence_id = ?',
          ['shipping', now, transactionEvidence.id]
        );

        break;

      case 'shipping':
        const statusReq: ShipmentStatusReq = {
          reserve_id: shipping['reserve_id'],
        };

        const shipmentURL = await getShipmentServiceURL(conn);
        let statusRes;
        try {
          statusRes = await fetch(`${shipmentURL}/status`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${SHIPMENT_SERVICE_ISUCARI_API_KEY}`,
            },
            body: JSON.stringify(statusReq),
          });
        } catch (err) {
          await conn.rollback();
          return { error: 'failed to get shipment status' };
        }

        if (!statusRes.ok) {
          await conn.rollback();
          return { error: 'failed to get shipment status' };
        }

        const shipmentStatus = await statusRes.json() as ShipmentStatusRes;

        if (shipmentStatus.status !== 'shipping' && shipmentStatus.status !== 'done') {
          await conn.rollback();
          return { error: 'shipment service error' };
        }

        if (shipmentStatus.status === 'shipping') {
          await conn.rollback();
          return { error: 'item is still shipping' };
        }

        await conn.execute(
          'UPDATE shippings SET status = ?, updated_at = ? WHERE transaction_evidence_id = ?',
          ['done', now, transactionEvidence.id]
        );

        break;

      case 'done':
        await conn.rollback();
        return { error: 'item has already arrived' };

      default:
        await conn.rollback();
        return { error: 'unknown shipping status' };
    }

    await conn.execute(
      'UPDATE transaction_evidences SET status = ?, updated_at = ? WHERE id = ?',
      ['wait_done', now, transactionEvidence.id]
    );

    await conn.commit();
    return { transactionEvidenceID: transactionEvidence.id };

  } catch (err) {
    await conn.rollback();
    throw err;
  }
}

export async function complete(
  conn: PoolConnection,
  buyerID: number,
  itemID: number
): Promise<{ transactionEvidenceID: number } | { error: string }> {
  await conn.beginTransaction();

  try {
    const [items] = await conn.query<MySQLResultRows>(
      'SELECT * FROM items WHERE id = ? FOR UPDATE',
      [itemID]
    );

    if (items.length === 0) {
      await conn.rollback();
      return { error: 'item not found' };
    }

    const item = items[0] as Item;

    if (item.buyer_id !== buyerID) {
      await conn.rollback();
      return { error: 'forbidden' };
    }

    const [evidences] = await conn.query<MySQLResultRows>(
      'SELECT * FROM transaction_evidences WHERE item_id = ? FOR UPDATE',
      [itemID]
    );

    if (evidences.length === 0) {
      await conn.rollback();
      return { error: 'transaction evidence not found' };
    }

    const transactionEvidence = evidences[0] as TransactionEvidence;

    if (transactionEvidence.buyer_id !== buyerID) {
      await conn.rollback();
      return { error: 'forbidden' };
    }

    if (transactionEvidence.status !== 'wait_done') {
      await conn.rollback();
      return { error: 'item is not waiting for completion' };
    }

    const [shippings] = await conn.query<MySQLResultRows>(
      'SELECT * FROM shippings WHERE transaction_evidence_id = ? FOR UPDATE',
      [transactionEvidence.id]
    );

    if (shippings.length === 0) {
      await conn.rollback();
      return { error: 'shipping not found' };
    }

    const shipping = shippings[0]!;

    // Check shipment service status
    const shipmentServiceURL = await getShipmentServiceURL(conn);
    const statusReq: ShipmentStatusReq = {
      reserve_id: shipping['reserve_id'],
    };

    let statusRes;
    try {
      statusRes = await fetch(`${shipmentServiceURL}/status`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${SHIPMENT_SERVICE_ISUCARI_API_KEY}`,
        },
        body: JSON.stringify(statusReq),
      });
    } catch (err) {
      await conn.rollback();
      return { error: 'failed to request to shipment service' };
    }

    if (!statusRes.ok) {
      await conn.rollback();
      return { error: 'failed to request to shipment service' };
    }

    const shipmentStatus = await statusRes.json() as ShipmentStatusRes;

    if (shipmentStatus.status !== 'done') {
      await conn.rollback();
      return { error: 'shipment service側で配送完了になっていません' };
    }

    const now = new Date();

    // Update shipping status to done
    await conn.execute(
      'UPDATE shippings SET status = ?, updated_at = ? WHERE transaction_evidence_id = ?',
      ['done', now, transactionEvidence.id]
    );

    await conn.execute(
      'UPDATE transaction_evidences SET status = ?, updated_at = ? WHERE id = ?',
      ['done', now, transactionEvidence.id]
    );

    await conn.execute(
      'UPDATE items SET status = ?, updated_at = ? WHERE id = ?',
      ['sold_out', now, itemID]
    );

    await conn.execute(
      'UPDATE users SET num_sell_items = num_sell_items + 1 WHERE id = ?',
      [item.seller_id]
    );

    await conn.commit();
    return { transactionEvidenceID: transactionEvidence.id };

  } catch (err) {
    await conn.rollback();
    throw err;
  }
}
