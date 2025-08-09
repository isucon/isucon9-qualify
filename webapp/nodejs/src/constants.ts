export const SESSION_NAME = 'session_isucari';

export const DEFAULT_PAYMENT_SERVICE_URL = 'http://localhost:5555';
export const DEFAULT_SHIPMENT_SERVICE_URL = 'http://localhost:7001';

export const ITEM_MIN_PRICE = 100;
export const ITEM_MAX_PRICE = 1000000;
export const ITEM_PRICE_ERR_MSG = '商品価格は100ｲｽｺｲﾝ以上、1,000,000ｲｽｺｲﾝ以下にしてください';

export const ITEM_STATUS = {
  ON_SALE: 'on_sale',
  TRADING: 'trading',
  SOLD_OUT: 'sold_out',
  STOP: 'stop',
  CANCEL: 'cancel',
} as const;

export const PAYMENT_SERVICE_ISUCARI_API_KEY = 'a15400e46c83635eb181-946abb51ff26a868317c';
export const PAYMENT_SERVICE_ISUCARI_SHOP_ID = '11';
export const SHIPMENT_SERVICE_ISUCARI_API_KEY = '75ugk2m37a750fwir5xr-22l6h4wmue1bwrubzwd0';

export const TRANSACTION_EVIDENCE_STATUS = {
  WAIT_SHIPPING: 'wait_shipping',
  WAIT_DONE: 'wait_done',
  DONE: 'done',
} as const;

export const SHIPPINGS_STATUS = {
  INITIAL: 'initial',
  WAIT_PICKUP: 'wait_pickup',
  SHIPPING: 'shipping',
  DONE: 'done',
} as const;

export const BUMP_CHARGE_SECONDS = 3 * 1000; // 3 seconds in milliseconds

export const ITEMS_PER_PAGE = 48;
export const TRANSACTIONS_PER_PAGE = 10;

export const BCRYPT_COST = 10;
