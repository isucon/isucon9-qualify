import React from 'react';

import { Avatar, Typography, TextField, Button, Grid, makeStyles } from '@material-ui/core';
import { LockOutlined } from '@material-ui/icons';
import { Link as RouteLink } from 'react-router-dom';

const useStyles = makeStyles(theme => ({
    avatar: {
        margin: theme.spacing(1),
        backgroundColor: theme.palette.secondary.main,
    },
    form: {
        width: '100%',
        marginTop: theme.spacing(1),
    },
    submit: {
        margin: theme.spacing(3, 0, 2),
    },
}));

interface SignInFormComponentProps {
    userId: string,
    password: string,
}

const SignInPageFormComponent: React.FC<SignInFormComponentProps> = ({ userId, password }) => {
    const classes = useStyles();

    return (
        <div>
            <Avatar className={classes.avatar}>
                <LockOutlined/>
            </Avatar>
            <Typography component="h1" variant="h5">
                ログインページ
            </Typography>
            <form className={classes.form} noValidate>
                <TextField
                    variant="outlined"
                    margin="normal"
                    required
                    fullWidth
                    id="id"
                    label="ログインID"
                    name="id"
                    autoFocus
                    value={userId}
                />
                <TextField
                    variant="outlined"
                    margin="normal"
                    required
                    fullWidth
                    id="password"
                    label="パスワード"
                    name="password"
                    type="password"
                    autoComplete="current-password"
                    value={password}
                />
                <Button
                    type="submit"
                    fullWidth
                    variant="contained"
                    color="primary"
                    className={classes.submit}
                >
                    ログイン
                </Button>
                <Grid container>
                    <Grid item xs>
                        <RouteLink to="/signup">新規登録はこちら</RouteLink>
                    </Grid>
                </Grid>
            </form>
        </div>
    );
};

export { SignInPageFormComponent }
