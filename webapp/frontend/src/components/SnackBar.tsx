import * as React from 'react';
import { Snackbar } from '@material-ui/core';
import IconButton from '@material-ui/core/IconButton';
import CloseIcon from '@material-ui/icons/Close';
import makeStyles from '@material-ui/core/styles/makeStyles';

const useStyles = makeStyles(theme => ({
  close: {
    padding: theme.spacing(0.5),
  },
}));

type Props = {
  open: boolean;
  message?: string;
  handleClose: (event: React.MouseEvent) => void;
};

const SnackBar: React.FC<Props> = ({ open, message, handleClose }) => {
  const classes = useStyles();

  const handleOnClose = (event: React.SyntheticEvent, reason: string) => {
    return handleClose(event as React.MouseEvent);
  };

  return (
    <React.Fragment>
      <Snackbar
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'left',
        }}
        open={open}
        autoHideDuration={6000}
        onClose={handleOnClose}
        message={<span>{message}</span>}
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
    </React.Fragment>
  );
};

export { SnackBar };
