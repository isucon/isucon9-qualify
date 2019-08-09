import AppClient from "../httpClients/appClient";
import PaymentClient from "../httpClients/paymentClient";
import { ThunkAction, ThunkDispatch } from "redux-thunk";
import { FormErrorState } from "../reducers/formErrorReducer";
import { push } from "connected-react-router";
import { Action, AnyAction } from "redux";
import { BuyReq } from "../types/appApiTypes";
import { routes } from "../routes/Route";
import { CardReq, CardRes } from "../types/paymentApiTypes";
import { PaymentResponseError } from "../errors/PaymentResponseError";
import { AppResponseError } from "../errors/AppResponseError";
import { ResponseError } from "../errors/ResponseError";

export const BUY_START = "BUY_START";
export const BUY_SUCCESS = "BUY_SUCCESS";
export const BUY_FAIL = "BUY_FAIL";
export const USING_CARD_FAIL = "USING_CARD_FAIL";

type State = void;
type ThunkResult<R> = ThunkAction<R, State, undefined, AnyAction>;

export function buyItemAction(
  itemId: number,
  cardNumber: string
): ThunkResult<void> {
  return (dispatch: ThunkDispatch<any, any, AnyAction>) => {
    Promise.resolve(() => {
      dispatch(buyStartAction());
    })
      .then(() => {
        return PaymentClient.post("/card", {
          card_number: cardNumber,
          shop_id: "11" // TODO getting from /settings
        } as CardReq);
      })
      .then((response: Response) => {
        if (!response.ok) {
          throw new PaymentResponseError(
            "request to /card of payment service was failed",
            response
          );
        }

        return response.json();
      })
      .catch((err: Error) => {
        // Wrapping to judge kinds of error
        throw new PaymentResponseError(err.message);
      })
      .then((body: CardRes) => {
        return AppClient.post("/buy", {
          item_id: itemId,
          token: body.token
        } as BuyReq);
      })
      .then((response: Response) => {
        if (!response.ok) {
          throw new AppResponseError(
            "request to /buy of app was failed",
            response
          );
        }

        return response.json();
      })
      .then(() => {
        dispatch(buySuccessAction());
        dispatch(push(routes.buyComplete.path));
      })
      .catch((err: Error) => {
        if (err instanceof ResponseError) {
          const res = err.getResponse();
          let action: Function;

          if (err instanceof PaymentResponseError) {
            action = usingCardFailAction;
          } else if (err instanceof AppResponseError) {
            action = buyFailAction;
          } else {
            action = buyFailAction;
          }

          if (res) {
            return res.json().then((body: any) => {
              if (body && body.error) {
                dispatch(action(body.error));
                return;
              }

              dispatch(action(err.message));
            });
          }

          dispatch(action(err.message));
          return;
        }

        dispatch(buyFailAction(err.message));
      });
  };
}

export interface BuyStartAction extends Action<typeof BUY_START> {}

export function buyStartAction(): BuyStartAction {
  return {
    type: BUY_START
  };
}

export interface BuySuccessAction extends Action<typeof BUY_SUCCESS> {}

export function buySuccessAction(): BuySuccessAction {
  return {
    type: BUY_SUCCESS
  };
}

export interface UsingCardFailAction extends Action<typeof USING_CARD_FAIL> {
  payload: FormErrorState;
}

export function usingCardFailAction(error: string): UsingCardFailAction {
  return {
    type: USING_CARD_FAIL,
    payload: {
      error: undefined,
      buyFormError: {
        cardError: error
      }
    }
  };
}
export interface BuyFailAction extends Action<typeof BUY_FAIL> {
  payload: FormErrorState;
}

export function buyFailAction(error: string): BuyFailAction {
  return {
    type: BUY_FAIL,
    payload: {
      error: undefined,
      buyFormError: {
        buyError: error
      }
    }
  };
}
