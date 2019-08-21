import React from 'react';
import { Redirect, Route, RouteProps } from 'react-router';
import { routes } from '../../routes/Route';
import LoadingComponent from '../LoadingComponent';

type Props = {
  isLoggedIn: boolean;
  loading: boolean;
  load: () => void;
  alreadyLoaded: boolean;
} & RouteProps;

const AuthRoute: React.FC<Props> = ({
  component: Component,
  isLoggedIn,
  loading,
  load,
  alreadyLoaded,
  ...rest
}) => {
  if (!Component) {
    throw new Error('component attribute required for AuthRoute component');
  }

  if (!alreadyLoaded) {
    load();
  }

  return (
    <Route
      {...rest}
      render={props =>
        loading ? (
          <LoadingComponent />
        ) : isLoggedIn ? (
          <Component {...props} />
        ) : (
          <Redirect to={routes.login.path} />
        )
      }
    />
  );
};

export { AuthRoute };
