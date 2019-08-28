import React from 'react';
import Fab from '@material-ui/core/Fab/Fab';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { Edit } from '@material-ui/icons';

const useStyles = makeStyles(theme => ({
  fab: {
    margin: theme.spacing(1),
    position: 'fixed',
    top: 'auto',
    bottom: '30px',
    right: '30px',
    width: '100px',
    height: '100px',
  },
}));

interface SellingButtonComponentProps {
  onClick: (e: React.MouseEvent) => void;
}

const SellingButtonComponent: React.FC<SellingButtonComponentProps> = ({
  onClick,
}) => {
  const classes = useStyles();

  return (
    <Fab className={classes.fab} color="primary" onClick={onClick}>
      <Edit fontSize="large" />
    </Fab>
  );
};

export { SellingButtonComponent };
