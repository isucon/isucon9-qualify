import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { FormErrorState } from '../reducers/formErrorReducer';
import { push } from 'connected-react-router';
import { AnyAction } from 'redux';
import { SellReq, SellRes } from '../types/appApiTypes';
import { routes } from '../routes/Route';

export const SELLING_ITEM_SUCCESS = 'SELLING_ITEM_SUCCESS';
export const SELLING_ITEM_FAIL = 'SELLING_ITEM_FAIL';

type State = void;
type ThunkResult<R> = ThunkAction<R, State, undefined, AnyAction>;

export function listItemAction(
  name: string,
  description: string,
  price: number,
  categoryId: number,
  image: Blob,
): ThunkResult<void> {
  return (dispatch: ThunkDispatch<any, any, AnyAction>) => {
    const body = new FormData();
    body.append('name', name);
    body.append('description', description);
    body.append('price', price.toString());
    body.append('category_id', categoryId.toString());
    body.append('image', 'TODO: image');
    AppClient.postFormData('/sell', body)
      .then((response: Response) => {
        if (!response.ok) {
          throw new Error('HTTP status not 200');
        }
        return response.json();
      })
      .then((body: SellRes) => {
        dispatch(sellingSuccessAction(body.id));
        dispatch(push(routes.top.path)); // TODO
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

export interface SellingSuccessAction {
  type: typeof SELLING_ITEM_SUCCESS;
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

export interface SellingFailAction {
  type: typeof SELLING_ITEM_FAIL;
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
