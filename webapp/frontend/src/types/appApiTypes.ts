interface CsrfRequiredReq {
    csrf_token: string
}

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
 * POST /sell
 */
// Request
export interface SellReq extends CsrfRequiredReq{
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
}
