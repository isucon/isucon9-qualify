import * as React from 'react';
import { Snackbar, Theme } from '@material-ui/core';
import IconButton from '@material-ui/core/IconButton';
import CloseIcon from '@material-ui/icons/Close';
import CheckCircleIcon from '@material-ui/icons/CheckCircle';
import ErrorIcon from '@material-ui/icons/Error';
import makeStyles from '@material-ui/core/styles/makeStyles';
import SnackbarContent from '@material-ui/core/SnackbarContent';

const useStyles = makeStyles((theme: Theme) => ({
  text: {
    display: 'flex',
    alignItems: 'center',
  },
  close: {
    padding: theme.spacing(0.5),
  },
  icon: {
    fontSize: 20,
    marginRight: theme.spacing(1),
  },
  success: {
    backgroundColor: theme.palette.secondary.main,
  },
  error: {
    backgroundColor: theme.palette.primary.main,
  },
}));

export type SnackBarVariant = 'success' | 'error';

type Props = {
  open: boolean;
  variant: SnackBarVariant;
  message?: string;
  handleClose: (event: React.MouseEvent) => void;
};

const getIcon = (
  variant: SnackBarVariant,
): typeof CheckCircleIcon | typeof ErrorIcon => {
  switch (variant) {
    case 'success':
      return CheckCircleIcon;
    case 'error':
      return ErrorIcon;
    default:
      return CheckCircleIcon;
  }
};

const SnackBar: React.FC<Props> = ({ open, variant, message, handleClose }) => {
  const classes = useStyles();

  const handleOnClose = (event: React.SyntheticEvent, reason: string) => {
    return handleClose(event as React.MouseEvent);
  };
  const Icon = getIcon(variant);

  return (
    <Snackbar
      anchorOrigin={{
        vertical: 'bottom',
        horizontal: 'left',
      }}
      open={open}
      autoHideDuration={6000}
      onClose={handleOnClose}
    >
      <SnackbarContent
        className={classes[variant]}
        message={
          <span className={classes.text}>
            <Icon className={classes.icon} />
            {message}
          </span>
        }
        action={[
          <IconButton
            key="close"
            aria-label="close"
            color="inherit"
            className={classes.close}
            onClick={handleClose}
          >
            <CloseIcon />
          </IconButton>,
        ]}
      />
    </Snackbar>
  );
};

export { SnackBar };
