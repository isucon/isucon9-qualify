import * as React from "react";
import { Button, createStyles, Theme, WithStyles } from "@material-ui/core";
import CircularProgress from "@material-ui/core/CircularProgress";
import { StyleRules } from "@material-ui/core/styles";
import withStyles from "@material-ui/core/styles/withStyles";

const styles = (theme: Theme): StyleRules =>
  createStyles({
    button: {
      margin: theme.spacing(3, 0, 1)
    },
    buttonProgress: {
      position: "absolute",
      top: "50%",
      left: "50%",
      marginTop: -12,
      marginLeft: -12
    }
  });

interface Props extends WithStyles<typeof styles> {
  onClick: (e: React.MouseEvent) => void;
  buttonName: string;
  loading: boolean;
}

class LoadingButtonComponent extends React.Component<Props> {
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
      <React.Fragment>
        <Button
          className={classes.button}
          disabled={loading}
          onClick={this._onClick}
        >
          {buttonName}
        </Button>
        {loading && (
          <CircularProgress size={24} className={classes.buttonProgress} />
        )}
      </React.Fragment>
    );
  }
}

export default withStyles(styles)(LoadingButtonComponent);
