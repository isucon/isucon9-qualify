import {ItemStatus} from "../dataObjects/item";
import {UserData} from "../dataObjects/user";

/**
 * POST /register
 */
export interface RegisterReq {
    account_name: string
    address: string
    password: string
}
export interface RegisterRes extends Response{
    id: number
    account_name: string
    address: string
}

/**
 * POST /login
 */
export interface LoginRes {
    id: number,
    account_name: string,
    address?: string,
    num_sell_items: number,
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

/**
 * GET /new_item.json
 */
export interface NewItemReq {
    item_id?: number,
    created?: number,
}

export interface NewItemRes {
    root_category_id?: number,
    root_category_name?: string,
    has_next: boolean,
    items: {
        id: number
        seller_id: number
        seller: {
            id: number,
            account_name: string,
            num_sell_items: number,
        },
        status: ItemStatus,
        name: string,
        price: number,
        category_id: number,
        category: {
            id: number,
            parent_id: number,
            category_name: string,
            parent_category_name: string,
        },
        created_at: number,
    }[]
}
