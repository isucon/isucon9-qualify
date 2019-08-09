import React from 'react';
import { Route, Switch } from 'react-router-dom';
import SignInPage from '../pages/SignInPage';
import SignUpPage from '../pages/SignUpPage';
import SellPage from '../pages/SellPage';
import ItemEditPage from '../pages/ItemEditPage';
import ItemBuyPage from '../pages/ItemBuyPage';
import UserSettingPage from '../pages/UserSettingPage';
import BuyCompletePage from '../pages/BuyComplete';
import ItemPageContainer from '../containers/ItemPageContainer';
import ItemListPageContainer from '../containers/ItemListPageContainer';
import TransactionPageContainer from '../containers/TransactionPageContainer';
import UserPageContainer from '../containers/UserPageContainer';
import AuthRoute from '../containers/AuthContainer';
import NonAuthRoute from '../containers/NonAuthContainer';
import NotFoundPage from '../pages/error/NotFoundPage';
import TopPage from '../pages/TopPage';

interface route {
  [name: string]: {
    path: string;
    getPath: (...params: any) => string; // TODO: optionalにしたい
  };
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
  timeline: {
    path: '/timeline',
    getPath: () => '/timeline',
  },
  categoryTimeline: {
    path: '/categories/:category_id/items',
    getPath: (categoryId: number) => `/categories/${categoryId}/items`,
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
    getPath: (itemId: number) => `/items/${itemId}/edit`,
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
      <NonAuthRoute exact path={routes.top.path} component={TopPage} />
      <NonAuthRoute exact path={routes.login.path} component={SignInPage} />
      <NonAuthRoute exact path={routes.register.path} component={SignUpPage} />
      <AuthRoute
        exact
        path={routes.timeline.path}
        component={ItemListPageContainer}
      />
      <AuthRoute exact path={routes.sell.path} component={SellPage} />
      <AuthRoute exact path={routes.item.path} component={ItemPageContainer} />
      <AuthRoute exact path={routes.itemEdit.path} component={ItemEditPage} />
      <AuthRoute exact path={routes.buy.path} component={ItemBuyPage} />
      <AuthRoute
        exact
        path={routes.buyComplete.path}
        component={BuyCompletePage}
      />
      <AuthRoute
        exact
        path={routes.transaction.path}
        component={TransactionPageContainer}
      />
      <AuthRoute exact path={routes.user.path} component={UserPageContainer} />
      <AuthRoute
        exact
        path={routes.userSetting.path}
        component={UserSettingPage}
      />
      <Route component={NotFoundPage} />
    </Switch>
  );
};
