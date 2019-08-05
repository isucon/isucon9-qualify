import { AnyAction } from "redux";
import {ItemData} from "../dataObjects/item";
import {FETCH_ITEM_PAGE_SUCCESS} from "../actions/fetchItemPageAction";

export interface ViewingItemState {
    item?: ItemData
}

const initialState: ViewingItemState = {
};

const viewingItem = (state: ViewingItemState = initialState, action: AnyAction): ViewingItemState => {
    switch (action.type) {
        case FETCH_ITEM_PAGE_SUCCESS:
            return {...state};
        default:
            return initialState;
    }
};

export default viewingItem;