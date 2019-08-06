import React from 'react';
import makeStyles from "@material-ui/core/styles/makeStyles";
import SignUpFormContainer from "../containers/SignUpFormContainer";
import {BasePageComponent} from "../components/BasePageComponent";

const useStyles = makeStyles(theme => ({
    paper: {
        marginTop: theme.spacing(1),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
}));

const SignUpPage: React.FC = () => {
    const classes = useStyles();

    return (
        <BasePageComponent>
            <div className={classes.paper}>
                <SignUpFormContainer />
            </div>
        </BasePageComponent>
    );
};

export default SignUpPage;