import React from 'react';

import { Typography, TextField, Button, Grid, makeStyles } from '@material-ui/core';
import { Link as RouteLink } from 'react-router-dom';
import Paper from "@material-ui/core/Paper";
import Avatar from "@material-ui/core/Avatar";
import {Camera} from "@material-ui/icons";

const useStyles = makeStyles(theme => ({
    paper: {
        marginTop: theme.spacing(1),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
    form: {
        width: '80%',
        marginTop: theme.spacing(1),
    },
    submit: {
        margin: theme.spacing(3, 0, 2),
    },
    upload: {
        display: 'none',
    },
    button: {
        margin: theme.spacing(1),
    },
    avatar: {
        margin: theme.spacing(1),
    },
}));

const SellPage: React.FC = () => {
    const classes = useStyles();

    return (
        <div className={classes.paper}>
            <Typography component="h1" variant="h5">
                出品ページ
            </Typography>
            <form className={classes.form} noValidate>
                <Grid
                    container
                    direction="row"
                    justify="space-between"
                    alignItems="center"
                >
                    <Grid item xs={8}>
                        <Paper>
                            <Avatar className={classes.avatar}>
                                <Camera />
                            </Avatar>
                            <Typography>商品画像</Typography>
                        </Paper>
                    </Grid>
                    <Grid item xs={4}>
                        <input
                            accept="image/*"
                            className={classes.upload}
                            id="outlined-button-file"
                            multiple
                            type="file"
                        />
                        <label htmlFor="outlined-button-file">
                            <Button variant="outlined" component="span" className={classes.button}>
                                Upload
                            </Button>
                        </label>
                    </Grid>
                </Grid>

                <TextField
                    variant="outlined"
                    margin="normal"
                    required
                    fullWidth
                    id="name"
                    label="商品名"
                    name="name"
                    autoFocus
                />
                <TextField
                    variant="outlined"
                    margin="normal"
                    required
                    fullWidth
                    id="description"
                    label="商品説明"
                    name="name"
                    multiline
                />
                <TextField
                    variant="outlined"
                    margin="normal"
                    required
                    fullWidth
                    id="price"
                    label="値段"
                    name="price"
                />
                <Button
                    type="submit"
                    fullWidth
                    variant="contained"
                    color="primary"
                    className={classes.submit}
                >
                    出品する
                </Button>
            </form>
        </div>
    );
};

export { SellPage }
