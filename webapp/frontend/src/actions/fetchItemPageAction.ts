import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from "redux-thunk";
import {Action, AnyAction} from "redux";
import {GetItemRes} from "../types/appApiTypes";
import {AppResponseError} from "../errors/AppResponseError";
import {ItemData} from "../dataObjects/item";
import {NotFoundError} from "../errors/NotFoundError";

export const FETCH_ITEM_PAGE_START = 'FETCH_ITEM_PAGE_START';
export const FETCH_ITEM_PAGE_SUCCESS = 'FETCH_ITEM_PAGE_SUCCESS';
export const FETCH_ITEM_PAGE_FAIL = 'FETCH_ITEM_PAGE_FAIL';

type ThunkResult<R> = ThunkAction<R, void, undefined, AnyAction>

export function fetchItemPageAction(itemId: string): ThunkResult<void> {
    return (dispatch: ThunkDispatch<any, any, AnyAction>) => {
        Promise.resolve(() => {
            dispatch(fetchItemPageStartAction());
        })
            .then(() => AppClient.get(`/items/${itemId}.json`))
            .then((response: Response) => {
                if (!response.ok) {
                    if (response.status === 404) {
                        throw new NotFoundError('Item not found');
                    }

                    throw new AppResponseError('Request for getting item data was failed', response);
                }

                return response.json();
            })
            .then((body: GetItemRes) => {
                dispatch(fetchItemPageSuccessAction({
                    id: body.id,
                    status: body.status,
                    sellerId: body.seller_id,
                    name: body.name,
                    price: body.price,
                    thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png', // TODO
                    description: body.description,
                    createdAt: '2019-07-20 12:00:00', // TODO
                }));
            })
            .catch((err: Error) => {
                dispatch(fetchItemPageFailAction());
            });
    };
}

export interface FetchItemPageStartAction extends Action<typeof FETCH_ITEM_PAGE_START> {}

const fetchItemPageStartAction = (): FetchItemPageStartAction => {
    return {
        type: FETCH_ITEM_PAGE_START,
    };
};

export interface FetchItemPageSuccessAction extends Action<typeof FETCH_ITEM_PAGE_SUCCESS > {
    payload: {
        item: ItemData,
    },
}

const fetchItemPageSuccessAction = (item: ItemData): FetchItemPageSuccessAction => {
    return {
        type: FETCH_ITEM_PAGE_SUCCESS ,
        payload: {
            item
        },
    };
};

export interface FetchItemPageFailAction extends Action<typeof FETCH_ITEM_PAGE_FAIL > {}

const fetchItemPageFailAction = (): FetchItemPageFailAction => {
    return {
        type: FETCH_ITEM_PAGE_FAIL ,
    };
};

