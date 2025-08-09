export interface Config {
  name: string;
  val: string;
}

export interface User {
  id: number;
  account_name: string;
  hashed_password: Buffer;
  address: string;
  num_sell_items: number;
  last_bump: Date;
  created_at: Date;
}

export interface UserSimple {
  id: number;
  account_name: string;
  num_sell_items: number;
}

export interface Item {
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
}

export interface ItemSimple {
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
}

export interface ItemDetail {
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
}

export interface TransactionEvidence {
  id: number;
  seller_id: number;
  buyer_id: number;
  status: string;
  item_id: number;
  item_name: string;
  item_price: number;
  item_description: string;
  item_category_id: number;
  item_root_category_id: number;
  created_at: Date;
  updated_at: Date;
}

export interface Shipping {
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
  img_binary: Buffer;
  created_at: Date;
  updated_at: Date;
}

export interface Category {
  id: number;
  parent_id: number;
  category_name: string;
  parent_category_name?: string;
}

export interface ReqInitialize {
  payment_service_url: string;
  shipment_service_url: string;
}

export interface ReqRegister {
  account_name: string;
  address: string;
  password: string;
}

export interface ReqLogin {
  account_name: string;
  password: string;
}

export interface ResNewItems {
  root_category_id?: number;
  root_category_name?: string;
  has_next: boolean;
  items: ItemSimple[];
}

export interface ResUserItems {
  user: UserSimple;
  has_next: boolean;
  items: ItemSimple[];
}

export interface ResTransactions {
  has_next: boolean;
  items: ItemDetail[];
}

export interface ReqItemEdit {
  csrf_token: string;
  item_id: number;
  item_price: number;
}

export interface ReqBuy {
  csrf_token: string;
  item_id: number;
  token: string;
}

export interface ReqSell {
  csrf_token: string;
  name: string;
  description: string;
  price: number;
  category_id: number;
}

export interface ReqShip {
  csrf_token: string;
  item_id: number;
}

export interface ReqShipDone {
  csrf_token: string;
  item_id: number;
}

export interface ReqComplete {
  csrf_token: string;
  item_id: number;
}

export interface ReqBump {
  csrf_token: string;
  item_id: number;
}

export interface ResInitialize {
  campaign: number;
  language: string;
}

export interface ResSettings {
  csrf_token: string;
  user: User | null;
  categories: Category[];
  payment_service_url: string;
}

export interface ResTransactionEvidence {
  transaction_evidence_id: number;
}

export interface ResItemEdit {
  item_id: number;
  item_price: number;
  item_created_at: number;
  item_updated_at: number;
}

export interface ResSell {
  id: number;
}

export interface ResShip {
  path: string;
  reserve_id: string;
}

export interface ResBump {
  item_id: number;
  item_price: number;
  item_created_at: number;
  item_updated_at: number;
}

export interface ResLogin {
  id: number;
  account_name: string;
  address: string;
}

export interface ResRegister {
  id: number;
  account_name: string;
  address: string;
}

export interface ResError {
  error: string;
}
