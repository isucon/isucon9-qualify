/**
 * POST /register
 */
// Request
export interface RegisterReqParams {
    accountName: string
    address: string
    password: string
}

// Response
export interface RegisterResParams extends Response{
    id: number
    account_name: string
    address: string
}
