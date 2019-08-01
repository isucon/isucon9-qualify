import React from 'react';

import { makeStyles } from '@material-ui/core';
import {SignInPageFormComponent} from "../components/SignInFormComponent";

const useStyles = makeStyles(theme => ({
    paper: {
        marginTop: theme.spacing(1),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
}));

const SignInPage: React.FC = () => {
    const classes = useStyles();

    return (
        <div className={classes.paper}>
            <SignInPageFormComponent userId={"1235"} password={"password"}/>
        </div>
    );
};

export { SignInPage }