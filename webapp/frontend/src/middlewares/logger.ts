import { Dispatch, Middleware, MiddlewareAPI } from 'redux';
import { AppState } from '../index';
import { ActionTypes } from '../actions/actionTypes';

const logger: Middleware = <S extends AppState>({
  getState,
}: MiddlewareAPI<Dispatch, S>) => (next: Dispatch<ActionTypes>) => (
  action: ActionTypes,
): any => {
  console.group(action.type);
  console.info('dispatching', action);
  const result = next(action);
  console.log('next state', getState());
  console.groupEnd();
  return result;
};

export default logger;
