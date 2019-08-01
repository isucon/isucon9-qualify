import {combineReducers, compose} from 'redux';
import authStatus from './authStatusReducer';

const reducers = combineReducers({
    authStatus,
});

export default reducers;