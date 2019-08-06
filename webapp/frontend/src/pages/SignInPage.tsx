import React from 'react';

import { makeStyles } from '@material-ui/core';
import SignInFormContainer from "../containers/SignInFormContainer";
import {BasePageComponent} from "../components/BasePageComponent";

const useStyles = makeStyles(theme => ({
    paper: {
        marginTop: theme.spacing(1),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
}));

type Props = {};

const SignInPage: React.FC<Props> = () => {
    const classes = useStyles();

    return (
        <BasePageComponent>
            <div className={classes.paper}>
                <SignInFormContainer />
            </div>
        </BasePageComponent>
    );
};

export default SignInPage;