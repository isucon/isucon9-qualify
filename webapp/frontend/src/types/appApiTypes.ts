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
