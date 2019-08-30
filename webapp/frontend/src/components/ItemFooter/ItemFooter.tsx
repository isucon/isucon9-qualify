import { AppBar, Theme } from '@material-ui/core';
import makeStyles from '@material-ui/core/styles/makeStyles';
import Grid from '@material-ui/core/Grid';
import * as React from 'react';
import Typography from '@material-ui/core/Typography';
import Button from '@material-ui/core/Button';

const useStyles = makeStyles((theme: Theme) => ({
  appBar: {
    top: 'auto',
    bottom: 0,
    padding: theme.spacing(0, 2),
  },
  buyButton: {
    margin: theme.spacing(1),
  },
}));

type Props = {
  price: number;
  buttons: {
    onClick: (e: React.MouseEvent) => void;
    buttonText: string;
    disabled: boolean;
  }[];
};

const ItemFooter: React.FC<Props> = ({ price, buttons }) => {
  const classes = useStyles();

  return (
    <AppBar color="primary" position="fixed" className={classes.appBar}>
      <Grid
        container
        spacing={2}
        direction="row"
        justify="space-between"
        alignItems="center"
      >
        <Grid item>
          <Typography variant="h5">{price}ｲｽｺｲﾝ</Typography>
        </Grid>
        <Grid item>
          <Grid container direction="row">
            {buttons.map(button => {
              return (
                <Grid item>
                  <Button
                    variant="contained"
                    className={classes.buyButton}
                    color="secondary"
                    onClick={button.onClick}
                    disabled={button.disabled}
                  >
                    {button.buttonText}
                  </Button>
                </Grid>
              );
            })}
          </Grid>
        </Grid>
      </Grid>
    </AppBar>
  );
};

export { ItemFooter };
