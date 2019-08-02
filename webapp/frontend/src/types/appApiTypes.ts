interface CsrfRequiredReq {
    csrf_token: string
}

/**
 * POST /register
 */
// Request
export interface RegisterReqParams {
    account_name: string
    address: string
    password: string
}
// Response
export interface RegisterResParams extends Response{
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
export interface SellRes {
    id: number,
}
