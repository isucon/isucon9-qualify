import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
import { Provider } from 'react-redux';
import { createBrowserHistory } from 'history';
import { ConnectedRouter } from 'connected-react-router';
import { getStore } from './configureStore';
import createRootReducer from './reducers/index';

const history = createBrowserHistory();
const rootReducers = createRootReducer(history);
const store = getStore(rootReducers, history);

export type AppState = ReturnType<typeof rootReducers>;

ReactDOM.render(
  <Provider store={store}>
    <ConnectedRouter history={history}>
      <App />
    </ConnectedRouter>
  </Provider>,
  document.getElementById('root'),
);
