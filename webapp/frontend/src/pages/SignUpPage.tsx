import React from 'react';
import makeStyles from "@material-ui/core/styles/makeStyles";
import SignUpFormContainer from "../containers/SignUpFormContainer";

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
        <div className={classes.paper}>
            <SignUpFormContainer />
        </div>
    );
};

export default SignUpPage;