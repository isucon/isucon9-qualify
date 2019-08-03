import {applyMiddleware, createStore, Store} from 'redux';
import createRootReducer from './reducers/index';
import { History } from 'history';
import { routerMiddleware } from "connected-react-router";
import thunk from 'redux-thunk';
import { composeWithDevTools } from 'redux-devtools-extension';
import middlewares from './middlewares';

export const getStore = (history: History): Store => {
    return createStore(
        createRootReducer(history),
        composeWithDevTools(
            applyMiddleware(
                thunk,
                routerMiddleware(history),
                ...middlewares,
            ),
        ),
    );
};
