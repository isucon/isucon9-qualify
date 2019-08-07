import {ItemStatus} from "../dataObjects/item";
import {UserData} from "../dataObjects/user";

/**
 * POST /register
 */
// Request
export interface RegisterReq {
    account_name: string
    address: string
    password: string
}
// Response
export interface RegisterRes extends Response{
    id: number
    account_name: string
    address: string
}

/**
 * GET /item
 */
export interface GetItemReq {
    item_id: number,
}
export interface GetItemRes {
    id: number,
    seller_id: number,
    seller: {
        id: number,
        account_name: string,
        num_sell_items: number,
    },
    buyer_id: number,
    buyer?: UserData,
    status: ItemStatus,
    name: string,
    price: number,
    description: string,
}

/**
 * POST /sell
 */
// Request
export interface SellReq {
    name: string,
    price: number,
    description: string,
}
// Response
export interface SellRes extends Response {
    id: number,
}

/**
 * GET /settings
 */
// Response
export interface SettingsRes {
    csrf_token: string,
    user?: {
        id: number,
        account_name: string,
        address?: string,
        num_sell_items: number,
    },
}

/**
 * POST /buy
 */
// Request
export interface BuyReq {
    item_id: number,
    token: string,
}

/**
 * Error response
 */
export interface ErrorRes {
    error: string,
}
