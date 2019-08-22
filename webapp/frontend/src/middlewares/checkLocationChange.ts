import { AnyAction, Dispatch, Middleware, MiddlewareAPI } from 'redux';
import { AppState } from '../index';
import { LOCATION_CHANGE, LocationChangeAction } from 'connected-react-router';
import { pathNameChangeAction } from '../actions/locationChangeAction';

// react-routerのページ遷移発火時、pathnameが変わった場合は独自のactionを発火する
const checkLocationChange: Middleware = <S extends AppState>(
  store: MiddlewareAPI<Dispatch, S>,
) => (next: Dispatch<AnyAction>) => (
  action: AnyAction | LocationChangeAction,
): any => {
  const { getState, dispatch } = store;
  if (action.type !== LOCATION_CHANGE) {
    return next(action);
  }

  const { router } = getState();
  const currentPath = router.location.pathname;
  const nextPath = action.payload.location.pathname;

  if (currentPath === nextPath) {
    return next(action);
  }

  dispatch(pathNameChangeAction());
  return next(action);
};

export default checkLocationChange;
