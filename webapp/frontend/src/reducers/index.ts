import { combineReducers } from 'redux';
import authStatus from './authStatusReducer';
import formError from './formErrorReducer';
import viewingItem from './viewingItemReducer';
import viewingUser from './viewingUserReducer';
import error from './errorReducer';
import page from './pageReducuer';
import snackBar from './snackBarReducer';
import buyPage from './buyPageReducer';
import categories from './categoriesReducer';
import timeline from './timelineReducer';
import transactions from './transactionsReducer';
import userItems from './userItemsReducer';
import { connectRouter } from 'connected-react-router';
import { History } from 'history';

export default (history: History) =>
  combineReducers({
    router: connectRouter(history),
    authStatus,
    formError,
    viewingItem,
    viewingUser,
    error,
    page,
    snackBar,
    buyPage,
    categories,
    timeline,
    transactions,
    userItems,
  });
