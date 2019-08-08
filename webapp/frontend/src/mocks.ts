import {ItemData} from "./dataObjects/item";
import {UserData} from "./dataObjects/user";

export const mockUser: UserData = {
    id: 1235,
    accountName: 'Kirin',
    address: 'Tokyo',
    numSellItems: 0,
};


export const mockItems: ItemData[] = [
    {
        id: 1,
        status: 'on_sale',
        sellerId: 1111,
        seller: {
            id: 1111,
            accountName: 'sota1235',
            address: "",
            numSellItems: 1,
        },
        buyerId: 2222,
        name: 'いす',
        price: 10000,
        description: 'いすです',
        thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
        category: {
            id: 1,
            categoryName: 'カテゴリ1',
            parentId: 2,
            parentCategoryName: '親カテゴリ',
        },
        createdAt: 10000,
    },
    {
        id: 2,
        status: 'on_sale',
        sellerId: 1111,
        seller: {
            id: 1111,
            accountName: 'sota1235',
            address: "",
            numSellItems: 1,
        },
        buyerId: 2222,
        name: 'いす',
        price: 10000,
        description: 'いすです',
        thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
        category: {
            id: 1,
            categoryName: 'カテゴリ1',
            parentId: 2,
            parentCategoryName: '親カテゴリ',
        },
        createdAt: 10000,
    },
    {
        id: 3,
        status: 'on_sale',
        sellerId: 1111,
        seller: {
            id: 1111,
            accountName: 'sota1235',
            address: "",
            numSellItems: 1,
        },
        buyerId: 2222,
        name: 'いす',
        price: 10000,
        description: 'いすです',
        thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
        category: {
            id: 1,
            categoryName: 'カテゴリ1',
            parentId: 2,
            parentCategoryName: '親カテゴリ',
        },
        createdAt: 10000,
    },
    {
        id: 4,
        status: 'on_sale',
        sellerId: 1111,
        seller: {
            id: 1111,
            accountName: 'sota1235',
            address: "",
            numSellItems: 1,
        },
        buyerId: 2222,
        name: 'いす',
        price: 10000,
        description: 'いすです',
        thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
        category: {
            id: 1,
            categoryName: 'カテゴリ1',
            parentId: 2,
            parentCategoryName: '親カテゴリ',
        },
        createdAt: 10000,
    },
];