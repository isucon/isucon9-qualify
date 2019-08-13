import { combineReducers } from 'redux';
import authStatus from './authStatusReducer';
import formError from './formErrorReducer';
import viewingItem from './viewingItemReducer';
import error from './errorReducer';
import page from './pageReducuer';
import buyPage from './buyPageReducer';
import categories from './categoriesReducer';
import timeline from './timelineReducer';
import { connectRouter } from 'connected-react-router';
import { History } from 'history';

export default (history: History) =>
  combineReducers({
    router: connectRouter(history),
    authStatus,
    formError,
    viewingItem,
    error,
    page,
    buyPage,
    categories,
    timeline,
  });
