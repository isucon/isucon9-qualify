import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
import { applyMiddleware, createStore } from 'redux';
import createRootReducer from './reducers/index';
import { Provider } from 'react-redux';
import { createBrowserHistory } from 'history';
import { ConnectedRouter, routerMiddleware } from "connected-react-router";
import thunk from 'redux-thunk';
import { composeWithDevTools } from 'redux-devtools-extension';
import middlewares from './middlewares';

const history = createBrowserHistory();

const store = createStore(
    createRootReducer(history),
    composeWithDevTools(
        applyMiddleware(
            thunk,
            routerMiddleware(history),
            ...middlewares,
        ),
    ),
);

ReactDOM.render(
    <Provider store={store}>
        <ConnectedRouter history={history}>
            <App />
        </ConnectedRouter>
    </Provider>,
    document.getElementById('root')
);
