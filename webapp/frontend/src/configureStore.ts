import { applyMiddleware, createStore, Reducer, Store } from 'redux';
import { History } from 'history';
import { routerMiddleware } from 'connected-react-router';
import thunk from 'redux-thunk';
import { composeWithDevTools } from 'redux-devtools-extension';
import middleware from './middlewares';

export const getStore = (reducer: Reducer, history: History): Store => {
  return createStore(
    reducer,
    composeWithDevTools(
      applyMiddleware(thunk, routerMiddleware(history), ...middleware),
    ),
  );
};
