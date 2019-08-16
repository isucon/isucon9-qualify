import React from 'react';
import { Link } from 'react-router-dom';
import './App.css';
import { AppRoute } from './routes/Route';

const routes: Array<{
  path: string;
  pageName: string;
}> = [
  {
    path: '/',
    pageName: 'Top page',
  },
  {
    path: '/login',
    pageName: 'Sign in page',
  },
  {
    path: '/register',
    pageName: 'Sign up page',
  },
  {
    path: '/timeline',
    pageName: 'Timeline page',
  },
  {
    path: '/items/:item_id/edit',
    pageName: 'Item edit page',
  },
  {
    path: '/users/:user_id',
    pageName: 'User page',
  },
  {
    path: '/users/setting',
    pageName: 'User setting page',
  },
];

const getLinks: () => React.ReactNode[] = () => {
  const routeComponents: React.ReactNode[] = [];

  for (const route of routes) {
    routeComponents.push(
      <li key={route.pageName}>
        <Link to={route.path}>{route.pageName}</Link>
      </li>,
    );
  }

  return routeComponents;
};

const App: React.FC<{}> = () => (
  <React.Fragment>
    <div>
      <ul>{getLinks()}</ul>
    </div>
    <hr />
    <AppRoute />
  </React.Fragment>
);

export default App;
