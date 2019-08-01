import React from 'react';
import { BrowserRouter as Router, Route, Link } from 'react-router-dom';

import './App.css';
import { TopPage } from "./pages/TopPage";
import { SignInPage } from "./pages/SignInPage";
import { SignUpPage } from "./pages/SignUpPage";
import { ItemListPage } from "./pages/ItemListPage";
import { ItemPage } from "./pages/ItemPage";
import { ItemEditPage } from "./pages/ItemEditPage";
import { SellPage } from "./pages/SellPage";
import { TransactionPage } from "./pages/TransactionPage";
import { UserPage } from "./pages/UserPage";
import { UserSettingPage } from "./pages/UserSettingPage";
import { BasePageComponent } from "./components/BasePageComponent";

const routes: Array<{
    path: string,
    component: any, // todo
    pageName: string,
}> = [
    {
        path: '/',
        component: TopPage,
        pageName: 'Top page'
    },
    {
        path: '/signin',
        component: SignInPage,
        pageName: 'Sign in page',
    },
    {
        path: '/signup',
        component: SignUpPage,
        pageName: 'Sign up page',
    }, {
        path: '/items',
        component: ItemListPage,
        pageName: 'Item list page',
    },
    {
        path: '/items/:item_id',
        component: ItemPage,
        pageName: 'Item page',
    },
    {
        path: '/items/:item_id/edit',
        component: ItemEditPage,
        pageName: 'Item edit page',
    },
    {
        path: '/sell',
        component: SellPage,
        pageName: 'Sell page',
    },
    {
        path: '/transactions/:transaction_id',
        component: TransactionPage,
        pageName: 'Transaction page',
    },
    {
        path: '/users/:user_id',
        component: UserPage,
        pageName: 'User page',
    },
    {
        path: '/users/:user_id/setting',
        component: UserSettingPage,
        pageName: 'User setting page',
    },
];

const getLinks: () => any[] = () => {
    const routeComponents: any[] = []; // TODO

    for (const route of routes) {
        routeComponents.push(
            <li>
                <Link to={route.path}>{route.pageName}</Link>
            </li>
        );
    }

    return routeComponents;
};

const getRoutes: () => any[] = () => {
    const routeComponents: any[] = []; // TODO

    for (const route of routes) {
        const component: React.FC = () => (
            <BasePageComponent>
                {route.component()}
            </BasePageComponent>
        );
        routeComponents.push(
            <Route exact path={route.path} component={component} />
        );
    }

    return routeComponents;
};

const App: React.FC = () => {
    return (
        <Router>
            <div>
                <ul>
                    {getLinks()}
                </ul>
            </div>

            <hr />

            {getRoutes()}
        </Router>
  );
};

export default App;
