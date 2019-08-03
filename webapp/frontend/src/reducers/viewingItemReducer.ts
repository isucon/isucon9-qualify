import { AnyAction } from "redux";
import {ItemData} from "../dataObjects/item";
import {FETCH_ITEM_START, FETCH_ITEM_FAIL, FETCH_ITEM_SUCCESS} from "../actions/fetchItemAction";

export interface ViewingItemState {
    item?: ItemData
    isFetching: boolean
}

const initialState: ViewingItemState = {
    isFetching: false,
};

const viewingItem = (state: ViewingItemState = initialState, action: AnyAction): ViewingItemState => {
    switch (action.type) {
        case FETCH_ITEM_START:
            return {...state, isFetching: true};
        case FETCH_ITEM_SUCCESS:
        case FETCH_ITEM_FAIL:
            return {...state, isFetching: false};
        default:
            return initialState;
    }
};

export default viewingItem;