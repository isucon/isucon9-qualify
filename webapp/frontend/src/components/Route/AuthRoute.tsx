import React from 'react';
import { Redirect, Route, RouteProps } from 'react-router';
import { routes } from '../../routes/Route';

type Props = {
  isLoggedIn: boolean;
} & RouteProps;

const AuthRoute: React.FC<Props> = ({
  component: Component,
  isLoggedIn,
  ...rest
}) => {
  if (!Component) {
    throw new Error('component attribute required for AuthRoute component');
  }

  return (
    <Route
      {...rest}
      render={props =>
        isLoggedIn ? (
          <Component {...props} />
        ) : (
          <Redirect to={routes.login.path} />
        )
      }
    />
  );
};

export { AuthRoute };
