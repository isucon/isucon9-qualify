import { AppBar, Theme } from '@material-ui/core';
import makeStyles from '@material-ui/core/styles/makeStyles';
import Grid from '@material-ui/core/Grid';
import * as React from 'react';
import Typography from '@material-ui/core/Typography';
import Button from '@material-ui/core/Button';
import Tooltip from '@material-ui/core/Tooltip';
import { ReactElement } from 'react';

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
    tooltip?: ReactElement;
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
              const ButtonComponent = (
                <Button
                  variant="contained"
                  className={classes.buyButton}
                  color="secondary"
                  onClick={button.onClick}
                  disabled={button.disabled}
                >
                  {button.buttonText}
                </Button>
              );

              // Hack: いいコードじゃないけど時間ないので許して
              if (button.tooltip) {
                return (
                  <Grid item>
                    <Tooltip title={button.tooltip} placement="top">
                      {ButtonComponent}
                    </Tooltip>
                  </Grid>
                );
              }

              return <Grid item>{ButtonComponent}</Grid>;
            })}
          </Grid>
        </Grid>
      </Grid>
    </AppBar>
  );
};

export { ItemFooter };
