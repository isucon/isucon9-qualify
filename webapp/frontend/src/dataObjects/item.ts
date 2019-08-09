import {UserData} from "./user";
import {Category} from "./category";

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
    category: Category,
    createdAt: number,
}

export type ItemStatus = 'on_sale' | 'trading' | 'sold_out' | 'stop' | 'cancel';
