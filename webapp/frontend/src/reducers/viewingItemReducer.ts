import { AnyAction } from "redux";
import {ItemData} from "../dataObjects/item";

export interface ViewingItemState {
    item?: ItemData
}

const initialState: ViewingItemState = {
};

const viewingItem = (state: ViewingItemState = initialState, action: AnyAction): ViewingItemState => {
    switch (action.type) {
        default:
            return initialState;
    }
};

export default viewingItem;