import React from 'react';
import Typography from "@material-ui/core/Typography/Typography";
import Grid from "@material-ui/core/Grid/Grid";

interface ErrorMessageComponentProps {
    errMsg: string[]
}

const ErrorMessageComponent: React.FC<ErrorMessageComponentProps> = ({ errMsg }) => {
    const errors = [];

    for (const error of errMsg) {
        errors.push(
            <Grid item xs>
                <Typography
                    key={error}
                    variant="body2"
                    color="error"
                >
                    {error}
                </Typography>
            </Grid>
        )
    }

    return (
        <Grid container>
            {errors}
        </Grid>
    );
};

export { ErrorMessageComponent }