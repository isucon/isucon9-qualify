import * as React from 'react';
import { Button, createStyles, Theme, WithStyles } from '@material-ui/core';
import { StyleRules } from '@material-ui/core/styles';
import withStyles from '@material-ui/core/styles/withStyles';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    root: {
      position: 'relative',
    },
    button: {
      margin: theme.spacing(1),
    },
  });

export interface Props extends WithStyles<typeof styles> {
  onClick: (e: React.MouseEvent) => void;
  buttonName: string;
  loading: boolean;
}

class BaseLoadingButton extends React.Component<Props> {
  constructor(props: Props) {
    super(props);

    this._onClick = this._onClick.bind(this);
  }

  _onClick(e: React.MouseEvent) {
    e.preventDefault();

    this.props.onClick(e);
  }

  render() {
    const { loading, buttonName, classes } = this.props;

    return (
      <div className={classes.root}>
        <Button
          className={classes.button}
          variant="contained"
          color="primary"
          disabled={loading}
          onClick={this._onClick}
        >
          {buttonName}
        </Button>
      </div>
    );
  }
}

export const LoadingButton = withStyles(styles)(BaseLoadingButton);
