import {UserData} from "./user";

export interface ItemData {
    id: number,
    sellerId: number,
    seller: UserData,
    buyerId: number,
    buyer?: UserData,
    status: ItemStatus,
    name: string,
    price: number,
    description: string,
    thumbnailUrl: string,
}

export type ItemStatus = 'on_sale' | 'trading' | 'sold_out' | 'stop' | 'cancel';
