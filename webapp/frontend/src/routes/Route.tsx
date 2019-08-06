import React from 'react';
import {Route, Switch} from "react-router-dom";
import SignInPage from "../pages/SignInPage";
import SignUpPage from "../pages/SignUpPage";
import SellPage from "../pages/SellPage";
import ItemEditPage from "../pages/ItemEditPage";
import ItemBuyPage from "../pages/ItemBuyPage";
import TransactionPage from "../pages/TransactionPage";
import UserPage from "../pages/UserPage";
import UserSettingPage from "../pages/UserSettingPage";
import BuyCompletePage from "../pages/BuyComplete";
import ItemListPage from "../pages/ItemListPage";
import ItemPageContainer from "../containers/ItemPageContainer";

interface route {
    [name: string]: {
        path: string,
        getPath: (...params: any) => string, // TODO: optionalにしたい
    }
}

export const routes: route = {
    top: {
        path: '/',
        getPath: () => '/',
    },
    login: {
        path: '/login',
        getPath: () => '/login',
    },
    register: {
        path: '/register',
        getPath: () => 'register',
    },
    sell: {
        path: '/sell',
        getPath: () => '/sell',
    },
    item: {
        path: '/items/:item_id',
        getPath: (itemId: number) => `/items/${itemId}`,
    },
    itemEdit: {
        path: '/items/:item_id/edit',
        getPath: (itemId: number) => `/items/${itemId}/edit`
    },
    buy: {
        path: '/items/:item_id/buy',
        getPath: (itemId: number) => `/items/${itemId}/buy`,
    },
    buyComplete: {
        path: '/buy/complete',
        getPath: () => '/buy/complete',
    },
    transaction: {
        path: '/transactions/:transaction_id',
        getPath: (transactionId: number) => `/transactions/${transactionId}`,
    },
    user: {
        path: '/users/:user_id',
        getPath: (userId: number) => `/users/${userId}`,
    },
    userSetting: {
        path: '/users/setting',
        getPath: () => '/users/setting',
    },
};

export const AppRoute: React.FC = () => {
    return (
        <Switch>
            <Route exact path={routes.top.path}         component={ItemListPage} />
            <Route exact path={routes.login.path}       component={SignInPage} />
            <Route exact path={routes.register.path}    component={SignUpPage}/>
            <Route exact path={routes.sell.path}        component={SellPage} />
            <Route exact path={routes.item.path}        component={ItemPageContainer} />
            <Route exact path={routes.itemEdit.path}    component={ItemEditPage} />
            <Route exact path={routes.buy.path}         component={ItemBuyPage} />
            <Route exact path={routes.buyComplete.path} component={BuyCompletePage} />
            <Route exact path={routes.transaction.path} component={TransactionPage} />
            <Route exact path={routes.user.path}        component={UserPage} />
            <Route exact path={routes.userSetting.path} component={UserSettingPage} />
        </Switch>
    );
};
