import React from 'react';
import Typography from "@material-ui/core/Typography/Typography";

interface ErrorMessageComponentProps {
    errMsg: string[]
}

const ErrorMessageComponent: React.FC<ErrorMessageComponentProps> = ({ errMsg }) => {
    const errors = [];

    for (const error of errMsg) {
        errors.push(
            <Typography
                key={error}
                variant="body2"
                color="error"
            >
                {error}
            </Typography>
        )
    }

    return (
        <React.Fragment>
            {errors}
        </React.Fragment>
    );
};

export { ErrorMessageComponent }