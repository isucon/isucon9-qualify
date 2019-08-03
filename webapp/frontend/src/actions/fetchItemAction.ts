import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from "redux-thunk";
import {Action} from "redux";
import {GetItemRes} from "../types/appApiTypes";
import {AppResponseError} from "../errors/AppResponseError";
import {ViewingItemState} from "../reducers/viewingItemReducer";
import {ItemData} from "../dataObjects/item";

export const FETCH_ITEM_START = 'FETCH_ITEM_START';
export const FETCH_ITEM_SUCCESS = 'FETCH_ITEM_SUCCESS';
export const FETCH_ITEM_FAIL = 'FETCH_ITEM_FAIL';

type State = void | ViewingItemState;
type FetchItemActions = FetchItemStartAction | FetchItemSuccessAction | FetchItemFailAction;
type ThunkResult<R> = ThunkAction<R, State, undefined, FetchItemActions>

export function fetchItemAction(itemId: string): ThunkResult<void> {
    return (dispatch: ThunkDispatch<any, any, FetchItemActions>) => {
        dispatch(fetchItemStartAction());
        AppClient.get(`/items/${itemId}.json`)
            .then((response: Response) => {
                if (!response.ok) {
                    if (response.status === 404) {
                        dispatch(fetchItemFailAction());
                        // TODO 404表示
                    }

                    throw new AppResponseError('Request for getting item data was failed', response);
                }

                return response.json();
            })
            .then((body: GetItemRes) => {
                dispatch(fetchItemSuccessAction({
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
                dispatch(fetchItemFailAction());
                // TODO handling error
            });
    };
}

export interface FetchItemStartAction extends Action<typeof FETCH_ITEM_START> {}

const fetchItemStartAction = (): FetchItemStartAction => {
    return {
        type: FETCH_ITEM_START,
    };
};

export interface FetchItemSuccessAction extends Action<typeof FETCH_ITEM_SUCCESS > {
    payload: {
        item: ItemData,
    },
}

const fetchItemSuccessAction = (item: ItemData): FetchItemSuccessAction => {
    return {
        type: FETCH_ITEM_SUCCESS ,
        payload: {
            item
        },
    };
};

export interface FetchItemFailAction extends Action<typeof FETCH_ITEM_FAIL > {}

const fetchItemFailAction = (): FetchItemFailAction => {
    return {
        type: FETCH_ITEM_FAIL ,
    };
};

