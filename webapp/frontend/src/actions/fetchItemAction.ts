import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from "redux-thunk";
import {AnyAction} from "redux";
import {GetItemRes} from "../types/appApiTypes";
import {AppResponseError} from "../errors/AppResponseError";
import {ViewingItemState} from "../reducers/viewingItemReducer";
import {ItemData} from "../dataObjects/item";

export const GET_ITEM_SUCCESS = 'GET_ITEM_SUCCESS';

type State = void | ViewingItemState;
type ThunkResult<R> = ThunkAction<R, State, undefined, AnyAction>

export function fetchItemAction(itemId: number): ThunkResult<void> {
    return (dispatch: ThunkDispatch<any, any, AnyAction>) => {
        AppClient.get(`/items/${itemId}.json`)
            .then((response: Response) => {
                if (!response.ok) {
                    if (response.status === 404) {
                        // TODO handle as not found
                    }

                    throw new AppResponseError('Request for getting item data was failed', response);
                }

                return response.json();
            })
            .then((body: GetItemRes) => {
                dispatch(getItemSuccessAction({
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
                // TODO handling error
            });
    };
}

export interface GetItemSuccessAction extends AnyAction {
    type: typeof GET_ITEM_SUCCESS,
    payload: {
        item: ItemData,
    },
}

const getItemSuccessAction = (item: ItemData): GetItemSuccessAction => {
    return {
        type: GET_ITEM_SUCCESS,
        payload: { item },
    };
};
