import { combineReducers } from 'redux';
import authStatus from './authStatusReducer';
import formError from './formErrorReducer';
import viewingItem from "./viewingItemReducer";
import { connectRouter } from 'connected-react-router';
import { History } from 'history';

export default (history: History) => combineReducers({
    router: connectRouter(history),
    authStatus,
    formError,
    viewingItem,
});
