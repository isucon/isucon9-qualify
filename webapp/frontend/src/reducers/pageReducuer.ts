import {AnyAction} from "redux";
import {FETCH_ITEM_PAGE_FAIL, FETCH_ITEM_PAGE_START, FETCH_ITEM_PAGE_SUCCESS} from "../actions/fetchItemPageAction";

export interface PageState {
    isLoading: boolean
}

const initialState: PageState = {
    isLoading: true,
};

const page = (state: PageState = initialState, action: AnyAction): PageState => {
    switch (action.type) {
        case FETCH_ITEM_PAGE_START:
            return { isLoading: true };
        case FETCH_ITEM_PAGE_SUCCESS:
        case FETCH_ITEM_PAGE_FAIL:
            return { isLoading: false };
        default:
            return {...state};
    }
};

export default page;