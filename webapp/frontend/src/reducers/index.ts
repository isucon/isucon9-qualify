import { combineReducers } from 'redux';
import authStatus from './authStatusReducer';
import { connectRouter } from 'connected-react-router';
import { History } from 'history';

export default (history: History) => combineReducers({
    router: connectRouter(history),
    authStatus,
});
