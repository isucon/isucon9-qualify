import React from 'react';
import { Redirect, Route, RouteProps } from 'react-router';
import { routes } from '../../routes/Route';
import LoadingComponent from '../LoadingComponent';
import { InternalServerErrorPage } from '../../pages/error/InternalServerErrorPage';

type Props = {
  isLoggedIn: boolean;
  loading: boolean;
  load: () => void;
  alreadyLoaded: boolean;
  error?: string;
} & RouteProps;

const AuthRoute: React.FC<Props> = ({
  component: Component,
  isLoggedIn,
  loading,
  load,
  alreadyLoaded,
  error,
  ...rest
}) => {
  if (!Component) {
    throw new Error('component attribute required for AuthRoute component');
  }

  if (error) {
    return <InternalServerErrorPage message={error} />;
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
