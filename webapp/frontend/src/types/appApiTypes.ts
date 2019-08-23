import { ItemStatus } from '../dataObjects/item';
import { UserData } from '../dataObjects/user';
import { TransactionStatus } from '../dataObjects/transaction';
import { ShippingStatus } from '../dataObjects/shipping';

type Category = {
  id: number;
  parent_id: number;
  category_name: string;
  parent_category_name: string;
};

type CategorySimple = {
  id: number;
  parent_id: number;
  category_name: string;
};

type User = {
  id: number;
  account_name: string;
  address?: string;
  num_sell_items: number;
};

type UserSimple = {
  id: number;
  account_name: string;
  num_sell_items: number;
};

export type ItemSimple = {
  id: number;
  seller_id: number;
  seller: UserSimple;
  status: ItemStatus;
  name: string;
  price: number;
  image_url: string;
  category_id: number;
  category: Category;
  created_at: number;
};

export type ItemDetail = {
  id: number;
  seller_id: number;
  seller: UserSimple;
  buyer_id?: number;
  buyer?: UserData;
  status: ItemStatus;
  name: string;
  price: number;
  description: string;
  image_url: string;
  category_id: number;
  category: Category;
  transaction_evidence_id?: number;
  transaction_evidence_status?: TransactionStatus;
  shipping_status?: ShippingStatus;
  created_at: number;
};

/**
 * POST /register
 */
export interface RegisterReq {
  account_name: string;
  address: string;
  password: string;
}
export interface RegisterRes extends Response {
  id: number;
  account_name: string;
  address: string;
  num_sell_items: number;
}

/**
 * POST /login
 */
export interface LoginRes {
  id: number;
  account_name: string;
  address?: string;
  num_sell_items: number;
}

/**
 * GET /item
 */
export interface GetItemRes extends ItemDetail {}

/**
 * POST /items/edit
 */
export interface ItemEditReq {
  item_id: number;
  item_price: number;
}

export interface ItemEditRes {
  item_id: number;
  item_price: number;
  item_created_at: number;
  item_updated_at: number;
}

/**
 * POST /sell
 */
export interface SellReq {
  name: string;
  price: number;
  description: string;
  category_id: number;
}

export interface SellRes extends Response {
  id: number;
}

/**
 * POST /bump
 */
export interface BumpReq {
  item_id: number;
}
export interface BumpRes extends ItemEditRes {}

/**
 * GET /settings
 */
export interface SettingsRes {
  csrf_token: string;
  payment_service_url: string;
  user?: User;
  categories: CategorySimple[];
}

/**
 * POST /buy
 */
export interface BuyReq {
  item_id: number;
  token: string;
}

/**
 * Error response
 */
export interface ErrorRes {
  error: string;
}

/**
 * GET /new_item.json
 */
export interface NewItemReq {
  item_id?: number;
  created_at?: number;
}

export interface NewItemRes {
  root_category_id?: number;
  root_category_name?: string;
  has_next: boolean;
  items: ItemSimple[];
}
/**
 * GET /new_item.json
 */
export interface NewCategoryItemReq extends NewItemReq {}

export interface NewCategoryItemRes {
  root_category_id?: number;
  root_category_name?: string;
  has_next: boolean;
  items: ItemSimple[];
}

/**
 * POST /ship
 */
export interface ShipReq {
  item_id: number;
}
export interface ShipRes {
  path: string;
}
/**
 * POST /ship_done
 */
export interface ShipDoneReq {
  item_id: number;
}
export interface ShipDoneRes {
  transaction_evidence_id: string;
}
/**
 * POST /complete
 */
export interface CompleteReq {
  item_id: number;
}
export interface CompleteRes {
  transaction_evidence_id: string;
}
/**
 * GET /users/transactions.json
 */
export interface UserTransactionsReq {
  item_id?: number;
  created_at?: number;
}
export interface UserTransactionsRes {
  has_next: boolean;
  items: ItemDetail[];
}
/**
 * GET /users/:user_id.json
 * ユーザの出品商品一覧
 */
export interface UserItemsReq {
  item_id?: number;
  created_at?: number;
}
export interface UserItemsRes {
  user: UserSimple;
  has_next: boolean;
  items: ItemSimple[];
}
