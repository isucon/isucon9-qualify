import React from 'react';
import makeStyles from "@material-ui/core/styles/makeStyles";
import { Avatar, Typography } from "@material-ui/core";
import { LockOutlined } from "@material-ui/icons";
import TextField from "@material-ui/core/TextField";
import Button from "@material-ui/core/Button";
import Grid from "@material-ui/core/Grid";
import { Link as RouteLink } from 'react-router-dom';

const useStyles = makeStyles(theme => ({
    paper: {
        marginTop: theme.spacing(1),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
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

const SignUpPage: React.FC = () => {
    const classes = useStyles();

    return (
        <div className={classes.paper}>
            <Avatar className={classes.avatar}>
                <LockOutlined/>
            </Avatar>
            <Typography component="h1" variant="h5">
                新規登録
            </Typography>
            <form className={classes.form} noValidate>
                <TextField
                    variant="outlined"
                    margin="normal"
                    required
                    fullWidth
                    id="name"
                    label="ユーザ名"
                    name="name"
                    autoFocus
                />
                <TextField
                    variant="outlined"
                    margin="normal"
                    required
                    fullWidth
                    id="address"
                    label="住所"
                    name="address"
                />
                <TextField
                    variant="outlined"
                    margin="normal"
                    required
                    fullWidth
                    id="id"
                    label="ログインID"
                    name="id"
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
                />
                <Button
                    type="submit"
                    fullWidth
                    variant="contained"
                    color="primary"
                    className={classes.submit}
                >
                    新規登録
                </Button>
                <Grid container>
                    <Grid item xs>
                        <RouteLink to="/signin">すでにアカウントをお持ちの方はこちら</RouteLink>
                    </Grid>
                </Grid>
            </form>
        </div>
    );
};

export { SignUpPage }