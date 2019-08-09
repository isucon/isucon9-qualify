import { AnyAction, Dispatch, Middleware, MiddlewareAPI } from 'redux';
import { AppState } from '../index';

const logger: Middleware = <S extends AppState>({
  getState,
}: MiddlewareAPI<Dispatch, S>) => (next: Dispatch<AnyAction>) => (
  action: any,
): any => {
  console.group(action.type);
  console.info('dispatching', action);
  const result = next(action);
  console.log('next state', getState());
  console.groupEnd();
  return result;
};

export default logger;
