import React from 'react';
import Paper from '@material-ui/core/Paper/Paper';
import Avatar from '@material-ui/core/Avatar/Avatar';
import { Camera } from '@material-ui/icons';
import Typography from '@material-ui/core/Typography/Typography';
import { Theme } from '@material-ui/core/styles/createMuiTheme';
import CloudUpload from '@material-ui/icons/CloudUpload';
import withStyles, {
  WithStyles,
  StyleRules,
} from '@material-ui/core/styles/withStyles';
import createStyles from '@material-ui/core/styles/createStyles';
import Grid from '@material-ui/core/Grid/Grid';
import Button from '@material-ui/core/Button/Button';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    imagePreviewSpace: {
      display: 'flex',
      justifyContent: 'center',
      flexDirection: 'column',
      alignItems: 'center',
      height: '300px',
    },
    imagePreviewText: {
      margin: theme.spacing(2, 0),
    },
    imagePreview: {
      width: 'auto',
      height: '100%',
      maxHeight: '300px',
    },
    upload: {
      display: 'none',
    },
    icon: {
      marginLeft: theme.spacing(1),
    },
    avatar: {
      margin: theme.spacing(1),
    },
    button: {
      margin: theme.spacing(1),
    },
  });

interface Props extends WithStyles<typeof styles> {
  onImageChange: (image: Blob) => void;
}

interface State {
  file?: File;
  imagePreviewUrl: string;
}

class ItemImageUploadComponent extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);

    this.state = {
      imagePreviewUrl: '',
    };
    this._handleImageChange = this._handleImageChange.bind(this);
  }

  _handleImageChange(e: React.ChangeEvent<HTMLInputElement>) {
    e.preventDefault();

    if (e.target.files === null) {
      return;
    }

    const reader = new FileReader();
    const file = e.target.files[0];

    reader.onloadend = () => {
      if (typeof reader.result !== 'string') {
        throw new Error();
      }

      this.setState({
        file: file,
        imagePreviewUrl: reader.result,
      });

      this.props.onImageChange(file);
    };

    reader.readAsDataURL(file);
  }

  render() {
    const { classes } = this.props;
    const { imagePreviewUrl } = this.state;
    let imagePreview = null;

    if (imagePreviewUrl) {
      imagePreview = (
        <img
          className={classes.imagePreview}
          alt="プレビュー"
          src={imagePreviewUrl}
        />
      );
    } else {
      imagePreview = (
        <Paper className={classes.imagePreviewSpace}>
          <Avatar className={classes.avatar}>
            <Camera />
          </Avatar>
          <Typography className={classes.imagePreviewText} variant="h6">
            商品画像
          </Typography>
        </Paper>
      );
    }

    return (
      <Grid
        container
        direction="row"
        justify="space-between"
        alignItems="center"
      >
        <Grid item xs={8}>
          {imagePreview}
        </Grid>
        <Grid item xs={4}>
          <input
            accept="image/*"
            className={classes.upload}
            id="outlined-button-file"
            type="file"
            onChange={this._handleImageChange}
          />
          <label htmlFor="outlined-button-file">
            <Button
              variant="outlined"
              color="default"
              component="span"
              className={classes.button}
            >
              Upload
              <CloudUpload className={classes.icon} />
            </Button>
          </label>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(ItemImageUploadComponent);
