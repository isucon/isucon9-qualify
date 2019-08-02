import React from 'react';
import makeStyles from "@material-ui/core/styles/makeStyles";
import SignUpFormComponent from "../components/SignUpFormComponent";

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
            <SignUpFormComponent register={
                (accountName: string, address: string, password: string) => {
                    // todo
                }
            }/>
        </div>
    );
};

export { SignUpPage }