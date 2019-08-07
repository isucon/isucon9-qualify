import React, {ReactElement} from 'react';
import {Redirect} from "react-router";
import {routes} from "../routes/Route";

type Props = {
    children: ReactElement
    isLoggedIn: boolean
}

const NonAuthComponent: React.FC<Props> = (props: Props) =>
    !props.isLoggedIn ? props.children : <Redirect to={routes.timeline.path} />;

export { NonAuthComponent }