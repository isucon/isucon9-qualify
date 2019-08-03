export interface ItemData {
    id: number,
    status?: ItemStatus,
    sellerId?: number,
    name: string,
    price: number,
    thumbnailUrl: string,
    description: string,
    createdAt: string,
}

export type ItemStatus = 'on_sale' | 'trading' | 'sold_out' | 'stop' | 'cancel';
