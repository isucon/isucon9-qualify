/**
 * POST /card
 */
export interface CardReq {
  card_number: string;
  shop_id: string;
}

export interface CardRes extends Response {
  token: string;
}
