import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { FormErrorState } from '../reducers/formErrorReducer';
import { CallHistoryMethodAction, push } from 'connected-react-router';
import { Action } from 'redux';
import { ErrorRes, SellRes } from '../types/appApiTypes';
import { routes } from '../routes/Route';
import { AppResponseError } from '../errors/AppResponseError';
import { AppState } from '../index';

export const SELLING_ITEM_SUCCESS = 'SELLING_ITEM_SUCCESS';
export const SELLING_ITEM_FAIL = 'SELLING_ITEM_FAIL';

export type SellingItemActions =
  | SellingSuccessAction
  | SellingFailAction
  | CallHistoryMethodAction;
type ThunkResult<R> = ThunkAction<R, AppState, undefined, SellingItemActions>;

export function listItemAction(
  name: string,
  description: string,
  price: number,
  categoryId: number,
  image: Blob,
): ThunkResult<void> {
  return (dispatch: ThunkDispatch<AppState, any, SellingItemActions>) => {
    const body = new FormData();
    body.append('name', name);
    body.append('description', description);
    body.append('price', price.toString());
    body.append('category_id', categoryId.toString());
    body.append('image', image);
    AppClient.postFormData('/sell', body)
      .then(async (response: Response) => {
        if (!response.ok) {
          const errRes: ErrorRes = await response.json();
          throw new AppResponseError(errRes.error, response);
        }
        return await response.json();
      })
      .then((body: SellRes) => {
        dispatch(sellingSuccessAction(body.id));
        dispatch(push(routes.timeline.path));
      })
      .catch((err: Error) => {
        dispatch(
          sellingFailAction({
            error: err.message,
          }),
        );
      });
  };
}

export interface SellingSuccessAction
  extends Action<typeof SELLING_ITEM_SUCCESS> {
  payload: {
    itemId: number;
  };
}

export function sellingSuccessAction(itemId: number): SellingSuccessAction {
  return {
    type: SELLING_ITEM_SUCCESS,
    payload: { itemId },
  };
}

export interface SellingFailAction extends Action<typeof SELLING_ITEM_FAIL> {
  payload: FormErrorState;
}

export function sellingFailAction(
  newErrors: FormErrorState,
): SellingFailAction {
  return {
    type: SELLING_ITEM_FAIL,
    payload: newErrors,
  };
}
