import React from 'react';
import { Redirect, Route, RouteProps } from 'react-router';
import { routes } from '../../routes/Route';

type Props = {
  isLoggedIn: boolean;
} & RouteProps;

const NonAuthRoute: React.FC<Props> = ({
  component: Component,
  isLoggedIn,
  ...rest
}) => {
  if (!Component) {
    throw new Error('component attribute required for NonAuthRoute component');
  }

  return (
    <Route
      {...rest}
      render={props =>
        !isLoggedIn ? (
          <Component {...props} />
        ) : (
          <Redirect to={routes.timeline.path} />
        )
      }
    />
  );
};

export { NonAuthRoute };
